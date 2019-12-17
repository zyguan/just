package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/pingcap/errors"
	"github.com/zyguan/just"
)

func load(p string) (m map[string]interface{}) {
	f := just.Try(os.Open(p)).Nth(0).(*os.File)
	defer f.Close()
	bs := just.TryTo("read " + p)(ioutil.ReadAll(f)).Nth(0).([]byte)
	just.TryTo("decode " + p)(json.Unmarshal(bs, &m))
	return
}

func printAll(ps []string) (err error) {
	defer just.Return(&err)
	for i, p := range ps {
		m := load(p)
		bs := just.Try(json.Marshal(m)).Nth(0).([]byte)
		log.Println(i, string(bs))
	}
	return
}

func main() {
	defer log.Println("# END")
	defer just.Catch(func(err error) {
		log.Printf("# OOPS: %+v", err)
	})
	log.Println("# BEGIN")
	just.Try(printAll(os.Args[1:]))
}

func init() {
	just.SetTraceFn(errors.Trace)
}
