package just

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var UncatchableErr = fmt.Errorf("oops: %s error", "uncatchable")
var CatchableErr = errors.New("should be catched")

func ok(val int) (int, error) {
	return val, nil
}

func fail(val int) (int, error) {
	return 0, UncatchableErr
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
		{ok, Wrap("ok"), 3},
		{fail, id, 2},
		{fail, Wrap("fail"), 4},
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
	for _, a := range []interface{}{UncatchableErr, "anwser", 42} {
		assert.NotPanics(t, func() {
			defer Catch(nil)
			Throw(a)
		})
		assert.Panics(t, func() {
			defer Catch(nil)
			panic(a)
		})
	}
	assert.NotPanics(t, func() {
		defer Catch(nil)
		Throwf("Life, the Universe and Everything: %d", 42)
	})
	assert.NotPanics(t, func() {
		defer Catch(nil)
		panic(CatchableErr)
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
		Throw(UncatchableErr)
	})
}

func TestCatchInnerError(t *testing.T) {
	defer CatchF(func(err error) error {
		assert.Equal(t, err.(WrappedCatcher).Cause(), UncatchableErr)
		return nil
	})(nil)
	foo := func() {
		// throw error in a func but not catch it
		Throw(UncatchableErr)
	}
	foo()
	assert.Fail(t, "shouldn't reach here")
}

func TestWrapError(t *testing.T) {
	for _, msg := range []string{"", "Hello", "World"} {
		assert.Equal(t, msg+": "+UncatchableErr.Error(), Wrap(msg)(UncatchableErr).Error())
		assert.Equal(t,
			errors.Wrap(UncatchableErr, msg).(WrappedCatcher).Cause(),
			Wrap(msg)(UncatchableErr).(WrappedCatcher).Cause())
	}
}

func TestTryTo(t *testing.T) {
	defer CatchF(func(err error) error {
		assert.True(t, strings.HasPrefix(err.Error(), "call fail: "))
		return nil
	})(nil)
	TryTo("call fail")(fail(0))
}

func TestHandleAll(t *testing.T) {
	defer HandleAll(func(err error) {
		assert.Equal(t, CatchableErr, err)
	})
	Throw(CatchableErr)
}
