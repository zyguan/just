package just

import "fmt"

// _Err is a type of catchable error
type _Err struct {
	error
}

func doCatch(ptr *error, f func(error) error, r interface{}) {
	switch r.(type) {
	case nil:
	case _Err:
		err := f(r.(_Err).error)
		if ptr != nil {
			*ptr = err
		}
	default:
		panic(r)
	}
}

// CatchF returns a function which assigns the captured error to its
// pointer argument.
func CatchF(f func(error) error) func(*error) {
	return func(ptr *error) { doCatch(ptr, f, recover()) }
}

// Catch is equivalent to CatchF(func(err error) error { return error }).
func Catch(ptr *error) {
	doCatch(ptr, id, recover())
}

// TryF returns a function which calls panic when an error occurs.
func TryF(f func(error) error) func(interface{}, error) interface{} {
	return func(val interface{}, err error) interface{} {
		if err != nil {
			panic(asErr(f(err)))
		}
		return val
	}
}

// Try is equivalent to TryF(func(err error) error { return error }).
func Try(val interface{}, err error) interface{} {
	return TryF(id)(val, err)
}

// Throw convert the argument to an catchable error and then panic it.
func Throw(a interface{}) {
	panic(asErr(a))
}

// Throwf just likes fmt.Errorf but panic the error directly.
func Throwf(format string, a ...interface{}) {
	panic(asErr(fmt.Errorf(format, a...)))
}

// Error just returns its err argument.
func Error(_ interface{}, err error) error {
	return err
}

func id(err error) error { return err }

func asErr(a interface{}) _Err {
	switch a.(type) {
	case error:
		return _Err{a.(error)}
	case string:
		return _Err{fmt.Errorf(a.(string))}
	default:
		return _Err{fmt.Errorf("%v", a)}
	}
}

// WithPrefix returns an error mapper which prepends the prefix to a
// given error. The error returned by the mapper is also an instance
// of ErrorWrapper.
func WithPrefix(pre string) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}
		return &_PreErr{pre, err}
	}
}

// TryTo is the same as TryF(WithPrefix(msg + ": "))
func TryTo(msg string) func(interface{}, error) interface{} {
	return TryF(WithPrefix(msg + ": "))
}

// ErrorWrapper is an error container. The internal error can be got
// by Err().
type ErrorWrapper interface {
	Err() error
}

type _PreErr struct {
	pre string
	err error
}

func (e *_PreErr) Error() string { return e.pre + e.err.Error() }

func (e *_PreErr) Err() error { return e.err }
