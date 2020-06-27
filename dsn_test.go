package gdsn

import (
	"testing"
)

func TestParse(t *testing.T) {
	d, err := Parse("wss://user::pas@s:word@0.0.0.0:8898/chat?timeout=3#frag")
	if err != nil {
		t.Error(err)
		return
	}

	if d.Address() != "0.0.0.0:8898" {
		t.Error("parse: incorrect address")
	}

	if pass, _ := d.User.Password(); pass != ":pas@s:word" {
		t.Error("parse: incorrect password")
	}
}

func TestDSN_Bind(t *testing.T) {
	d, err := Parse("unix:///var/run/mysql.sock?timeout=3&tag=a&tag=b")
	if err != nil {
		t.Fatal(err)
	}

	var config struct {
		Scheme  string `dsn:"scheme"`
		Address string `dsn:"address"`
		Query   struct {
			Timeout string `dsn:"timeout"`
		} `dsn:"query"`
		Tags    []string `dsn:"query.tag"`
		Timeout int      `dsn:"query.timeout"`
	}

	if err := d.Bind(&config); err != nil {
		t.Fatal(err)
	}

	if config.Scheme != "unix" {
		t.Fatal("bind: incorrect scheme")
	}

	if config.Address != "/var/run/mysql.sock" {
		t.Error("parse: incorrect address")
	}

	if config.Query.Timeout != "3" {
		t.Error("parse: incorrect timeout query inside struct")
	}

	if config.Timeout != 3 {
		t.Error("parse: incorrect timeout query")
	}
}
