package just

import (
	"fmt"

	"github.com/pkg/errors"
)

// Catcher is a type of catchable error
type Catcher interface {
	Error() string
	Format(s fmt.State, verb rune)
	StackTrace() errors.StackTrace
}

// WrappedCatcher is Catcher with an internal error
type WrappedCatcher interface {
	Catcher
	Cause() error
}

func doCatch(ptr *error, f func(error) error, r interface{}) {
	switch r.(type) {
	case nil:
	case Catcher:
		err := f(r.(Catcher))
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

// HandleAll deals with all catchable error.
func HandleAll(handle func(err error)) {
	doCatch(nil, func(err error) error {
		handle(err)
		return nil
	}, recover())
}

// TryF returns a function which panics when an error occurs.
func TryF(f func(error) error) func(interface{}, error) interface{} {
	return func(val interface{}, err error) interface{} {
		if err != nil {
			e := f(err)
			if _, ok := e.(Catcher); ok {
				panic(e)
			}
			panic(errors.WithStack(err))
		}
		return val
	}
}

// Try is equivalent to TryF(func(err error) error { return error }).
func Try(val interface{}, err error) interface{} {
	return TryF(id)(val, err)
}

// TryTo is the same as TryF(Wrap(msg)).
func TryTo(msg string) func(interface{}, error) interface{} {
	return TryF(Wrap(msg))
}

// Throw convert the argument to an catchable error and then panic it.
func Throw(a interface{}) {
	panic(catchable(a))
}

// Throwf just likes errors.Errorf but panic the error immediately.
func Throwf(format string, a ...interface{}) {
	panic(errors.Errorf(format, a...))
}

// Error just returns its err argument.
func Error(_ interface{}, err error) error {
	return err
}

// Wrap converts errors.Wrap to an error mapper.
func Wrap(msg string) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}
		return errors.Wrap(err, msg)
	}
}

func id(err error) error { return err }

func catchable(a interface{}) Catcher {
	switch a.(type) {
	case Catcher:
		return a.(Catcher)
	case error:
		return errors.WithStack(a.(error)).(Catcher)
	case string:
		return errors.New(a.(string)).(Catcher)
	default:
		return errors.Errorf("%v", a).(Catcher)
	}
}
