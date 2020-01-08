package main

import (
	"go.uber.org/zap"

	"github.com/pingcap/errors"
	"github.com/zyguan/just"
)

func f() (err error) {
	defer just.AnnotateAndReturn("some annotation")(&err)
	just.Try(errors.New("oops"))
	return
}

func main() {
	zap.S().Info(f())
	x := f()
	zap.S().Infof("%+v", x)
	zap.L().Error(">>>", zap.Error(f()))
}

func init() {
	just.SetTraceFn(errors.Trace)
	zap.ReplaceGlobals(just.Try(zap.NewDevelopment()).(*zap.Logger))
}
