// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/8

package xt

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type Testing interface {
	Fatalf(format string, args ...any)
}

type Helper interface {
	Helper()
}

var _ Testing = (*Collector)(nil)
var _ Helper = (*Collector)(nil)

// Collector 这个实现了 Testing ，可并发安全
type Collector struct {
	names  []string
	errors []string
	mux    sync.Mutex
}

func (t *Collector) Helper() {}

func (t *Collector) Run(name string, fn func(t Testing)) {
	nc := &Collector{
		names: slices.Clone(t.names),
	}
	nc.names = append(nc.names, name)
	fn(nc)
	errs := nc.getErrors()
	if len(errs) == 0 {
		return
	}
	fullName := strings.Join(nc.names, "/")
	t.fatalf("test=%q Fatal:\n %s", fullName, strings.Join(errs, "\n"))
}

func (t *Collector) getErrors() []string {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.errors
}

func (t *Collector) Fatalf(format string, args ...any) {
	t.fatalf(format, args...)
}

func (t *Collector) fatalf(format string, args ...any) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.errors = append(t.errors, fmt.Sprintf(format, args...))
}

func (t *Collector) Check(r Testing) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if len(t.errors) == 0 {
		return
	}
	txt := strings.Join(t.errors, "\n")
	r.Fatalf("%s", txt)
}
