package unit_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zyguan/just"
)

func TestError(t *testing.T) {
	must := just.Assert(t, require.New(t).NoError)
	must.Try(nil)
	must.Try(errors.New("oops"))
}
