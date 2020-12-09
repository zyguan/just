package just

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Catchable interface {
	Why() error
}

type caught struct{ err error }

func (c *caught) Why() error { return c.err }

func (c *caught) String() string { return c.err.Error() }

type withStack interface {
	HasStack() bool
}

type wrappedErr struct {
	msg       string
	err       error
	withStack bool
}

func wrap(err error, msg string) error {
	if len(msg) == 0 {
		return err
	}
	e := wrappedErr{msg: msg, err: err}
	_, e.withStack = err.(withStack)
	return &e
}

func (e *wrappedErr) Error() string { return e.msg + ": " + e.err.Error() }

func (e *wrappedErr) Cause() error { return e.err }

func (e *wrappedErr) HasStack() bool { return e.withStack }

func (e *wrappedErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", e.Cause())
			io.WriteString(s, e.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

func Return(ptr *error) {
	HandleRecovered(ptr, func(c Catchable) error { return c.Why() }, recover())
}

func HandleAndReturn(handle func(Catchable) error) func(*error) {
	return func(ptr *error) { HandleRecovered(ptr, handle, recover()) }
}

func AnnotateAndReturn(msg string) func(*error) {
	return HandleAndReturn(func(c Catchable) error { return wrap(c.Why(), msg) })
}

func Catch(handle func(c Catchable)) {
	HandleRecovered(nil, func(err Catchable) error {
		handle(err)
		return nil
	}, recover())
}

func HandleRecovered(ptr *error, handle func(Catchable) error, recovered interface{}) {
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

func Pack(xs ...interface{}) Values { return xs }

func (xs Values) Error() error { return ExtractError(xs...) }

func (xs Values) Nth(i int) interface{} {
	if i < 0 {
		i += len(xs)
	}
	if i < 0 || i >= len(xs) {
		return nil
	}
	return xs[i]
}

func ExtractError(xs ...interface{}) error {
	if len(xs) > 0 {
		if e, ok := xs[len(xs)-1].(error); ok {
			return e
		}
	}
	return nil
}

func AsCatchable(a interface{}, optMsgs ...string) Catchable {
	return trace.AsCatchable(a, optMsgs...)
}

func Try(xs ...interface{}) interface{} { return trace.Try(xs...).Nth(0) }

func TryTo(msg string) func(...interface{}) interface{} {
	return func(xs ...interface{}) interface{} {
		if err := ExtractError(xs...); err != nil {
			if c := trace.wrap(err, msg); c != nil {
				panic(c)
			}
		}
		return Values(xs).Nth(0)
	}
}

func TryValues(xs ...interface{}) Values { return trace.Try(xs...) }

func TryValuesWithMsg(msg string) func(...interface{}) Values {
	return func(xs ...interface{}) Values {
		if err := ExtractError(xs...); err != nil {
			if c := trace.wrap(err, msg); c != nil {
				panic(c)
			}
		}
		return xs
	}
}

func Throw(a interface{}) { trace.Throw(a) }

func Throwf(format string, args ...interface{}) { trace.Throwf(format, args...) }

type TraceFn func(error) error

func (f TraceFn) wrap(err error, optMsgs ...string) Catchable {
	if err == nil {
		return nil
	}
	tracedErr := f(err)
	if tracedErr == nil {
		return nil
	}
	if len(optMsgs) == 0 {
		return &caught{tracedErr}
	}
	return &caught{wrap(err, strings.Join(optMsgs, ": "))}
}

func (f TraceFn) AsCatchable(a interface{}, optMsgs ...string) Catchable {
	switch x := a.(type) {
	case Catchable:
		if len(optMsgs) == 0 {
			return x
		}
		return &caught{wrap(x.Why(), strings.Join(optMsgs, ": "))}
	case error:
		return f.wrap(x, optMsgs...)
	case string:
		return f.wrap(errors.New(x), optMsgs...)
	default:
		return f.wrap(fmt.Errorf("%v", a), optMsgs...)
	}
}

func (f TraceFn) Try(xs ...interface{}) Values {
	if err := ExtractError(xs...); err != nil {
		if c := f.wrap(err); c != nil {
			panic(c)
		}
	}
	return xs
}

func (f TraceFn) Throw(a interface{}) { panic(f.AsCatchable(a)) }

func (f TraceFn) Throwf(format string, args ...interface{}) { f.Throw(fmt.Errorf(format, args...)) }

func (f TraceFn) Errorf(format string, args ...interface{}) { f.Throwf(format, args...) }

var trace TraceFn = id

func id(err error) error { return err }

func SetTraceFn(f func(error) error) {
	if f == nil {
		f = id
	}
	trace = f
}
