// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/8

package xt

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
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

type TB interface {
	Run(name string, fn func(t TB))
	TestReporter
}

type TestReporter interface {
	Logf(format string, args ...any)
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
	FailNow()
}

var _ Testing = (*Collector)(nil)
var _ Helper = (*Collector)(nil)
var _ TB = (*Collector)(nil)

// Collector 这个实现了 Testing ，可并发安全
type Collector struct {
	parentNames []string // 父节点名字
	peerNames   []string // 兄弟节点名字
	logs        []logLine
	failed      bool
	mux         sync.Mutex
}

type logLine struct {
	log string
	by  logLineBy
}

type logLineBy string

const (
	logLineByFatalf logLineBy = "Fatalf"
	logLineByErrorf logLineBy = "Errorf"
	logLineByLogf   logLineBy = "Logf"
)

func (t *Collector) Failed() bool {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.failed
}

func (t *Collector) setFailed() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.failed = true
}

func (t *Collector) FailNow() {
	t.setFailed()
}

func (t *Collector) Helper() {}

func (t *Collector) Run(name string, fn func(t TB)) {
	t.mux.Lock()
	nc := &Collector{
		parentNames: slices.Clone(t.parentNames),
	}
	rawName := name
	for i := 1; ; i++ {
		if slices.Contains(t.peerNames, name) {
			name = fmt.Sprintf("%s#%d", rawName, i)
		} else {
			break
		}
	}
	nc.parentNames = append(nc.parentNames, name)
	t.peerNames = append(t.peerNames, name)
	t.mux.Unlock()

	fn(nc)

	caller := t.getCaller(2)
	fullName := strings.Join(nc.parentNames, "/")
	line := logLine{
		log: fmt.Sprintf("Run test=%q (%s)", fullName, caller),
		by:  "Run",
	}

	failed := nc.Failed()
	t.mux.Lock()
	defer t.mux.Unlock()

	t.logs = append(t.logs, line)
	if failed {
		t.failed = true
	}
	t.logs = append(t.logs, nc.getLogs()...)
}

func (t *Collector) getLogs() []logLine {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.logs
}

func (t *Collector) getCaller(skip int) string {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("%s:%d", filepath.Base(file), lineNo)
	}
	return fmt.Sprintf("%s:%d (%s)", filepath.Base(file), lineNo, path.Base(fn.Name()))
}

func (t *Collector) Errorf(format string, args ...any) {
	t.setFailed()
	t.log(fmt.Sprintf(format, args...), logLineByErrorf, 2)
}

func (t *Collector) Fatalf(format string, args ...any) {
	t.setFailed()
	t.log(fmt.Sprintf(format, args...), logLineByFatalf, 2)
}

func (t *Collector) Logf(format string, args ...any) {
	t.log(fmt.Sprintf(format, args...), logLineByLogf, 2)
}

func (t *Collector) log(txt string, by logLineBy, skip int) {
	lastCaller := t.getCaller(skip + 1)
	line := logLine{
		log: lastCaller + " " + txt,
		by:  by,
	}
	t.mux.Lock()
	defer t.mux.Unlock()
	t.logs = append(t.logs, line)
}

func (t *Collector) Check(r TestReporter) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if !t.failed {
		return
	}
	if th, ok := r.(Helper); ok {
		th.Helper()
	}
	for _, line := range t.logs {
		switch line.by {
		case logLineByFatalf:
			r.Fatalf("%s", line.log)
		case logLineByErrorf:
			r.Errorf("%s", line.log)
		case logLineByLogf:
			r.Logf("%s", line.log)
		default:
			r.Errorf("%s", line.log)
		}
	}
	r.FailNow()
}
