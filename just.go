package just

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Catchable interface {
	Why() error
	Error() string
}

func wrap(err error, optMsgs ...string) Catchable {
	if err == nil {
		return nil
	}
	msg := strings.Join(optMsgs, ": ")
	tracedErr := trace(err)
	if tracedErr == nil {
		tracedErr = err
	}
	return &caught{tracedErr, msg}
}

func AsCatchable(a interface{}, optMsgs ...string) Catchable {
	switch e := a.(type) {
	case Catchable:
		if len(optMsgs) == 0 {
			return e
		}
		return &caught{e.Why(), strings.Join(optMsgs, ": ")}
	case error:
		return wrap(e, optMsgs...)
	case string:
		return wrap(errors.New(e), optMsgs...)
	default:
		return wrap(fmt.Errorf("%v", a), optMsgs...)
	}
}

type caught struct {
	err error
	msg string
}

func (c *caught) Why() error { return c.err }

func (c *caught) Cause() error { return c.err }

func (c *caught) Error() string {
	if len(c.msg) == 0 {
		return c.err.Error()
	}
	return c.msg + ": " + c.err.Error()
}

func (c *caught) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", c.Why())
			if len(c.msg) > 0 {
				io.WriteString(s, "-- \n"+c.msg)
			}
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, c.Error())
	}
}

func Return(ptr *error) {
	HandleRecovered(ptr, id, recover())
}

func HandleAndReturn(handle func(error) error) func(*error) {
	return func(ptr *error) { HandleRecovered(ptr, handle, recover()) }
}

func AnnotateAndReturn(msg string) func(*error) {
	annotate := func(err error) error { return wrap(err, msg) }
	return HandleAndReturn(annotate)
}

func Catch(handle func(err error)) {
	HandleRecovered(nil, func(err error) error {
		handle(err)
		return nil
	}, recover())
}

func HandleRecovered(ptr *error, handle func(error) error, recovered interface{}) {
	if recovered == nil {
		return
	} else if x, ok := recovered.(Catchable); ok {
		err := handle(x)
		if ptr != nil {
			*ptr = err
		}
	} else {
		panic(recovered)
	}
}

type Values []interface{}

func (xs Values) Error() error { return ExtractError(xs) }

func (xs Values) Nth(i int) interface{} {
	if i < 0 {
		i += len(xs)
	}
	if i < 0 || i >= len(xs) {
		return nil
	}
	return xs[i]
}

func Try(xs ...interface{}) Values {
	if err := ExtractError(xs...); err != nil {
		panic(wrap(err))
	}
	return xs
}

func TryTo(msg string) func(...interface{}) Values {
	return func(xs ...interface{}) Values {
		if err := ExtractError(xs...); err != nil {
			panic(wrap(err, msg))
		}
		return xs
	}
}

func ExtractError(xs ...interface{}) error {
	if len(xs) > 0 {
		if e, ok := xs[len(xs)-1].(error); ok {
			return e
		}
	}
	return nil
}

func Throw(a interface{}) { panic(AsCatchable(a)) }

func Throwf(format string, args ...interface{}) { Throw(fmt.Errorf(format, args...)) }

type TraceFn func(error) error

var trace = id

func id(err error) error { return err }

func SetTraceFn(f func(error) error) {
	if f == nil {
		f = id
	}
	trace = f
}
