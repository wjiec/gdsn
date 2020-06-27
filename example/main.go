package main

import (
	"fmt"

	"github.com/wjiec/gdsn"
)

type config struct {
	Scheme  string `dsn:"scheme"`
	Address string `dsn:"address"`
	Options struct {
		Timeout int      `dsn:"timeout"`
		Tags    []string `dsn:"tag"`
	} `dsn:"query"`
	Limit int `dsn:"query.limit"`
}

func main() {
	d, err := gdsn.Parse("unix:///var/lib/wss.sock?timeout=3&tag=game&tag=any&limit=5")
	if err != nil {
		panic(err)
	}

	var cfg config
	if err := d.Bind(&cfg); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cfg)
}
