package main

import (
	"go.uber.org/zap"

	"github.com/pingcap/errors"
	"github.com/zyguan/just"
)

func f() (err error) {
	defer just.AnnotateAndReturn("defer annotation")(&err)
	just.Try(errors.New("oops"))
	return
}

func main() {
	zap.S().Info(f())
	zap.S().Infof("%+v", f())
	zap.L().Error(">>>", zap.Error(f()))
}

func init() {
	just.SetTraceFn(errors.Trace)
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l)
}
