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

type caught struct {
	err error
	msg string
}

func (c *caught) Why() error {
	if len(c.msg) > 0 {
		return c
	}
	return c.err
}

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
				io.WriteString(s, "\n"+c.msg)
			}
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, c.Error())
	}
}

func Return(ptr *error) {
	HandleRecovered(ptr, func(c Catchable) error { return c.Why() }, recover())
}

func HandleAndReturn(handle func(Catchable) error) func(*error) {
	return func(ptr *error) { HandleRecovered(ptr, handle, recover()) }
}

func AnnotateAndReturn(msg string) func(*error) {
	return HandleAndReturn(func(c Catchable) error { return &caught{c.Why(), msg} })
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

func Try(xs ...interface{}) Values { return trace.Try(xs...) }

func TryTo(msg string) func(...interface{}) Values { return trace.TryTo(msg) }

func Throw(a interface{}) { trace.Throw(a) }

func Throwf(format string, args ...interface{}) { trace.Throwf(format, args...) }

type TraceFn func(error) error

func (f TraceFn) wrap(err error, optMsgs ...string) Catchable {
	if err == nil {
		return nil
	}
	msg := strings.Join(optMsgs, ": ")
	tracedErr := f(err)
	if tracedErr == nil {
		tracedErr = err
	}
	return &caught{tracedErr, msg}
}

func (f TraceFn) AsCatchable(a interface{}, optMsgs ...string) Catchable {
	switch e := a.(type) {
	case Catchable:
		if len(optMsgs) == 0 {
			return e
		}
		return &caught{e.Why(), strings.Join(optMsgs, ": ")}
	case error:
		return f.wrap(e, optMsgs...)
	case string:
		return f.wrap(errors.New(e), optMsgs...)
	default:
		return f.wrap(fmt.Errorf("%v", a), optMsgs...)
	}
}

func (f TraceFn) Try(xs ...interface{}) Values {
	if err := ExtractError(xs...); err != nil {
		panic(f.wrap(err))
	}
	return xs
}

func (f TraceFn) TryTo(msg string) func(...interface{}) Values {
	return func(xs ...interface{}) Values {
		if err := ExtractError(xs...); err != nil {
			panic(f.wrap(err, msg))
		}
		return xs
	}
}

func (f TraceFn) Throw(a interface{}) { panic(f.AsCatchable(a)) }

func (f TraceFn) Throwf(format string, args ...interface{}) { f.Throw(fmt.Errorf(format, args...)) }

var trace TraceFn = id

func id(err error) error { return err }

func SetTraceFn(f func(error) error) {
	if f == nil {
		f = id
	}
	trace = f
}

type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

type ErrorAssertion func(err error, msgAndArgs ...interface{})

func Assert(t TestingT, optErrAssertions ...ErrorAssertion) TraceFn {
	if len(optErrAssertions) == 0 {
		optErrAssertions = append(optErrAssertions, func(err error, msgAndArgs ...interface{}) {
			if err != nil {
				t.Errorf("%+v", err)
				t.FailNow()
			}
		})
	}
	return func(e error) error {
		for _, assert := range optErrAssertions {
			assert(e, e.Error())
		}
		return nil
	}
}
