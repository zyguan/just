package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zyguan/just"
)

func openall(fs []string) (cnt int, err error) {
	defer just.CatchF(just.Wrap("openall failed"))(&err)
	for _, f := range fs {
		fd := just.TryTo(fmt.Sprintf("open file #%d:'%s'", cnt, f))(os.Open(f)).(*os.File)
		fmt.Printf("open %s as %#v\n", f, fd)
		cnt += 1
	}
	return cnt, nil
}

func main() {
	defer just.HandleAll(func(err error) {
		log.Fatal(err)
	})
	cnt := just.Try(openall(os.Args)).(int)
	fmt.Println("succeed in openning", cnt, "files!")
}
