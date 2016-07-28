package just

import "fmt"

type _Err struct {
	error
}

func CatchF(f func(error) error) func(*error) {
	return func(ptr *error) {
		switch r := recover(); r.(type) {
		case nil:
		case _Err:
			*ptr = f(r.(_Err).error)
		default:
			panic(r)
		}
	}
}

func Catch(ptr *error) {
	switch r := recover(); r.(type) {
	case nil:
	case _Err:
		*ptr = id(r.(_Err).error)
	default:
		panic(r)
	}
}

func TryF(f func(error) error) func(interface{}, error) interface{} {
	return func(val interface{}, err error) interface{} {
		if err != nil {
			panic(asErr(f(err)))
		}
		return val
	}
}

func Try(val interface{}, err error) interface{} {
	return TryF(id)(val, err)
}

func Throw(a interface{}) {
	panic(asErr(a))
}

func Throwf(format string, a ...interface{}) {
	panic(asErr(fmt.Errorf(format, a...)))
}

func Error(_ interface{}, err error) error {
	return err
}

func id(err error) error {
	return err
}

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

func WithPrefix(pre string) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}
		return &_PreErr{pre, err}
	}
}

type ErrorWrapper interface {
	Err() error
}

type _PreErr struct {
	pre string
	err error
}

func (e *_PreErr) Error() string {
	return e.pre + e.err.Error()
}

func (e *_PreErr) Err() error {
	return e.err
}
