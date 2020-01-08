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
	f := just.TryValues(os.Open(p)).Nth(0).(*os.File)
	defer f.Close()
	bs := just.TryValuesWithMsg("read " + p)(ioutil.ReadAll(f)).Nth(0).([]byte)
	just.TryValuesWithMsg("decode " + p)(json.Unmarshal(bs, &m))
	return
}

func printAll(ps []string) (err error) {
	defer just.Return(&err)
	for i, p := range ps {
		m := load(p)
		bs := just.TryValues(json.Marshal(m)).Nth(0).([]byte)
		log.Println(i, string(bs))
	}
	return
}

func main() {
	defer log.Println("# END")
	defer just.Catch(func(c just.Catchable) {
		log.Printf("# OOPS: %+v", c.Why())
	})
	log.Println("# BEGIN")
	just.TryValues(printAll(os.Args[1:]))
}

func init() {
	just.SetTraceFn(errors.Trace)
}
