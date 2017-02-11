package just

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var FakeErr = errors.New("fail")

func ok(val int) (int, error) {
	return val, nil
}

func fail(val int) (int, error) {
	return 0, FakeErr
}

func assertErr(t *testing.T, exp error, act error) {
	if exp != nil {
		assert.Equal(t, exp.Error(), act.Error())
	} else {
		assert.Nil(t, act)
	}
}

func TestTryCatch(t *testing.T) {
	wrap := func(
		f func(int) (int, error),
		try func(interface{}, error) interface{},
		catch func(ptr *error),
	) func(val int) (int, error) {
		return func(val int) (_ int, err error) {
			defer catch(&err)
			return try(f(val)).(int), nil
		}
	}

	for _, x := range []struct {
		f func(int) (int, error)
		e func(error) error
		v int
	}{
		{ok, id, 1},
		{ok, WithPrefix("ok: "), 3},
		{fail, id, 2},
		{fail, WithPrefix("fail: "), 4},
	} {
		expVal, expErr := x.f(x.v)

		actVal, actErr := wrap(x.f, Try, Catch)(x.v)
		assert.Equal(t, expVal, actVal)
		assertErr(t, expErr, actErr)

		actVal, actErr = wrap(x.f, Try, CatchF(x.e))(x.v)
		assert.Equal(t, expVal, actVal)
		assertErr(t, x.e(expErr), actErr)

		actVal, actErr = wrap(x.f, TryF(x.e), Catch)(x.v)
		assert.Equal(t, expVal, actVal)
		assertErr(t, x.e(expErr), actErr)

		actVal, actErr = wrap(x.f, TryF(x.e), CatchF(x.e))(x.v)
		assert.Equal(t, expVal, actVal)
		assertErr(t, x.e(x.e(expErr)), actErr)
	}
}

func TestThrowCatch(t *testing.T) {
	for _, a := range []interface{}{FakeErr, "anwser", 42} {
		assert.NotPanics(t, func() {
			var err error
			defer Catch(&err)
			Throw(a)
		})
		assert.Panics(t, func() {
			var err error
			defer Catch(&err)
			panic(a)
		})
	}
	assert.NotPanics(t, func() {
		var err error
		defer Catch(&err)
		Throwf("Life, the Universe and Everything: %d", 42)
	})
}

func TestCatchNil(t *testing.T) {
	var panicErr = func(err error) error {
		panic(err)
		return err
	}
	assert.NotPanics(t, func() {
		defer CatchF(panicErr)(nil)
	})
	assert.Panics(t, func() {
		defer CatchF(panicErr)(nil)
		Throw(FakeErr)
	})
}

func TestDeepCatch(t *testing.T) {
	defer CatchF(func(err error) error {
		assert.Equal(t, err, FakeErr)
		return nil
	})(nil)
	foo := func() {
		// throw error in a func but not catch it
		Throw(FakeErr)
	}
	foo()
	assert.Fail(t, "shouldn't reach here")
}

func TestPrefixError(t *testing.T) {
	for _, pre := range []string{"", "Hello", "World"} {
		assert.Equal(t, pre+FakeErr.Error(), WithPrefix(pre)(FakeErr).Error())
		assert.Equal(t, FakeErr, WithPrefix(pre)(FakeErr).(ErrorWrapper).Err())
	}
}

func TestTryTo(t *testing.T) {
	defer CatchF(func(err error) error {
		assert.True(t, strings.HasPrefix(err.Error(), "call fail: "))
		return nil
	})(nil)
	TryTo("call fail")(fail(0))
}
