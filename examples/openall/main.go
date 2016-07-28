package main

import (
	"fmt"
	"os"

	"github.com/zyguan/just"
)

func openall(fs []string) (cnt int, err error) {
	defer just.CatchF(just.WithPrefix("openall failed: "))(&err)
	for _, f := range fs {
		fd := just.TryF(just.WithPrefix("cannot open '" + f + "': "))(os.Open(f)).(*os.File)
		defer fd.Close()
		fmt.Printf("open %s as %#v\n", f, fd)
		cnt += 1
	}
	return cnt, nil
}

func main() {
	cnt, err := openall(os.Args)
	fmt.Println("open", cnt, "files, with error: ", err)
}
