package config

import (
	"strings"

	"github.com/sethvargo/go-envconfig"
)

type LookuperFunc func(key string) (string, bool)

func (f LookuperFunc) Lookup(key string) (string, bool) {
	return f(key)
}

type upcaseLookuper struct {
	Next envconfig.Lookuper
}

func (l *upcaseLookuper) Lookup(key string) (string, bool) {
	return l.Next.Lookup(strings.ToUpper(key))
}

func UpcaseLookuper(next envconfig.Lookuper) *upcaseLookuper {
	return &upcaseLookuper{
		Next: next,
	}
}
