package just

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func call(e error, f func(error) error) func(t *testing.T) {
	return func(t *testing.T) {
		if e != nil {
			assert.EqualError(t, f(e), e.Error())
		} else {
			assert.Nil(t, f(e))
		}
	}
}
func pcall(e error, f func(error) error, requireCatable bool) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Panics(t, func() {
			assert.NoError(t, f(e))
		})
		func() {
			defer func() {
				x := recover()
				if _, ok := x.(Catchable); requireCatable && !ok {
					assert.Fail(t, fmt.Sprintf("<%v: %T> is not a catchable error", x, x))
				}
			}()
			assert.NoError(t, f(e))
		}()
	}
}

func TestTryReturn(t *testing.T) {
	t.Run("return nil", call(nil, func(_ error) (err error) {
		defer Return(&err)
		Try(func() error { return nil }())
		return
	}))
	t.Run("return nil without catch", call(nil, func(e error) (err error) {
		Try(func() error { return nil }())
		return
	}))
	t.Run("return err", call(errors.New("an error"), func(e error) (err error) {
		defer Return(&err)
		Try(func() error { return e }())
		t.Fail()
		return
	}))
	t.Run("return err without catch", pcall(errors.New("an error"), func(e error) (err error) {
		Try(func() error { return e }())
		t.Fail()
		return
	}, true))

	t.Run("return val and nil", call(nil, func(_ error) (err error) {
		defer Return(&err)
		res := Try(func() (int, error) { return 42, nil }())
		assert.Equal(t, 42, res.Nth(0))
		return
	}))
	t.Run("return val and nil without catch", call(nil, func(e error) (err error) {
		res := Try(func() (int, error) { return 42, nil }())
		assert.Equal(t, 42, res.Nth(0))
		return
	}))
	t.Run("return val and err", call(errors.New("an error"), func(e error) (err error) {
		defer Return(&err)
		Try(func() (int, error) { return 42, e }())
		t.Fail()
		return
	}))
	t.Run("return val and err without catch", pcall(errors.New("an error"), func(e error) (err error) {
		Try(func() error { return e }())
		t.Fail()
		return
	}, true))
}

func TestThrowReturn(t *testing.T) {
	for _, a := range []interface{}{errors.New("an error"), "a string", 42, 3.14} {
		t.Run(fmt.Sprintf("throw %v", a), call(fmt.Errorf("%v", a), func(e error) (err error) {
			defer Return(&err)
			Throw(a)
			return
		}))
		t.Run(fmt.Sprintf("panic %v", a), pcall(nil, func(e error) (err error) {
			defer Return(&err)
			panic(a)
		}, false))
	}
	assert.NotPanics(t, func() {
		defer Return(nil)
		Throwf("Life, the Universe and Everything: %d", 42)
	})
	assert.NotPanics(t, func() {
		defer Return(nil)
		panic(wrap(errors.New("another error")))
	})
}

func TestAnnotateAndReturn(t *testing.T) {
	for _, a := range []interface{}{errors.New("an error"), "a string", 42, 3.14} {
		t.Run(fmt.Sprintf("annotate %v", a), call(fmt.Errorf("oops: %v", a), func(e error) (err error) {
			defer AnnotateAndReturn("oops")(&err)
			Throw(a)
			t.Fail()
			return
		}))
	}
}

func TestCatch(t *testing.T) {
	for _, a := range []interface{}{errors.New("an error"), "a string", 42, 3.14} {
		t.Run(fmt.Sprintf("catch %v", a), call(nil, func(e error) (err error) {
			defer Catch(func(err error) {
				assert.EqualError(t, err, fmt.Sprintf("%v", a))
			})
			Throw(a)
			t.Fail()
			return
		}))
	}
}

func TestCatchInnerError(t *testing.T) {
	e := errors.New("oops")

	defer Catch(func(err error) {
		c := err.(Catchable)
		assert.Equal(t, e, c.Why())
	})

	func() { func() { Throw(e) }() }()

	t.Fail()
}

func TestTryTo(t *testing.T) {
	t.Run("try to call func return nil", call(nil, func(e error) (err error) {
		defer Return(&err)
		TryTo("call func return nil")(func() error { return nil }())
		return
	}))
	t.Run("try to call func return err", call(errors.New("call func return err: oops"), func(e error) (err error) {
		defer Return(&err)
		TryTo("call func return err")(func() error { return errors.New("oops") }())
		return
	}))
}

func TestPanicInHandle(t *testing.T) {
	var panicInHandle = func(err error) error {
		panic(err)
	}
	assert.NotPanics(t, func() {
		defer HandleAndReturn(panicInHandle)(nil)
	})
	assert.Panics(t, func() {
		defer HandleAndReturn(panicInHandle)(nil)
		Throw("oops")
	})
}

func TestAsCatchable(t *testing.T) {
	c := wrap(errors.New("oops"))
	assert.Equal(t, c, AsCatchable(c))

	assert.Error(t, AsCatchable(nil))
	assert.Equal(t, nil, wrap(nil))
}

func TestNthValue(t *testing.T) {
	xs := Try(1, 2, 3)
	for i := -len(xs); i < len(xs); i++ {
		assert.NotNil(t, xs.Nth(i))
	}
	assert.Nil(t, xs.Nth(len(xs)))
	assert.Nil(t, xs.Nth(-len(xs)-1))
}

func TestExtractError(t *testing.T) {
	for _, tt := range []struct {
		xs     []interface{}
		hasErr bool
	}{
		{[]interface{}{}, false},
		{[]interface{}{1}, false},
		{[]interface{}{1, 2}, false},
		{[]interface{}{1, 2, errors.New("oops")}, true},
		{[]interface{}{errors.New("oops")}, true},
	} {
		if tt.hasErr {
			assert.Error(t, ExtractError(tt.xs...))
		} else {
			assert.NoError(t, ExtractError(tt.xs...))
		}
	}
}

func TestSetTraceFn(t *testing.T) {
	defer SetTraceFn(nil)
	type traced struct {
		error
		info string
	}
	trace := func(err error) error {
		return traced{err, "attached info"}
	}
	SetTraceFn(trace)
	e := errors.New("oops")
	c := wrap(e)
	te := c.Why().(traced)
	assert.Equal(t, e, te.error)
	assert.Equal(t, "attached info", te.info)
}

func BenchmarkJust(b *testing.B) {
	f := func() (err error) {
		defer Return(&err)
		x := Try(benchFn()).Nth(0).(float64)
		x += 1
		return
	}
	for i := 0; i < b.N; i++ {
		f()
	}
}

func BenchmarkIfError(b *testing.B) {
	f := func() error {
		x, err := benchFn()
		if err != nil {
			return err
		}
		x += 1
		return nil
	}
	for i := 0; i < b.N; i++ {
		f()
	}
}

var (
	benchErrRate = flag.Float64("err-rate", .1, "bench error rate")
	benchErr     = errors.New("bench error")
)

func benchFn() (float64, error) {
	x := rand.Float64()
	if x < *benchErrRate {
		return 0, benchErr
	}
	return x, nil
}
