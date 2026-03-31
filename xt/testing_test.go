// Copyright(C) 2024 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2024/4/25

package xt

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

func newMyTesting(t *testing.T) *myTesting {
	return &myTesting{
		t: t,
	}
}

var _ Testing = (*myTesting)(nil)
var _ Helper = (*myTesting)(nil)

type myTesting struct {
	t           *testing.T
	msg         string
	wantSuccess bool // 期望成功
	gotSuccess  bool // 是否运行成功
	lastCaller  string
}

func (m *myTesting) Helper() {
	if m.lastCaller != "" && m.gotSuccess != m.wantSuccess {
		m.check()
	}
	m.reset()
	_, file, lineNo, _ := runtime.Caller(2)
	m.lastCaller = fmt.Sprintf("%s:%d", filepath.Base(file), lineNo)
}

func (m *myTesting) reset() {
	m.gotSuccess = true
	m.msg = ""
	m.lastCaller = ""
}

func (m *myTesting) check() {
	defer m.reset()

	m.t.Helper()
	// m.t.Logf("%s wantSuccess=%v, gotSuccess=%v",m.lastCaller,m.wantSuccess,m.gotSuccess)
	if m.wantSuccess {
		if !m.gotSuccess {
			m.t.Fatalf("%s expect success, but not", m.lastCaller)
		}
	} else {
		if m.gotSuccess {
			m.t.Fatalf("%s expect fail, but not", m.lastCaller)
		}
	}
}

func (m *myTesting) Fatalf(format string, args ...any) {
	m.gotSuccess = false
	m.msg = fmt.Sprintf(format, args...)
}

func (m *myTesting) Success(fn func(t Testing)) {
	m.t.Helper()
	m.wantSuccess = true
	fn(m)
	m.check()
}

func (m *myTesting) Fail(fn func(t Testing)) {
	m.wantSuccess = false
	fn(m)
	m.check()
}

func TestCollector(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		tc := &Collector{}
		tc.Logf("ok")
		False(t, tc.Failed())

		tc.Fatalf("fail")
		True(t, tc.Failed())

		tc.Errorf("error")
		True(t, tc.Failed())

		logs := tc.getLogs()
		for _, line := range logs {
			t.Logf("line: %v", line)
			Contains(t, line.log, "xt.TestCollector")
			Contains(t, line.log, "testing_test.go")
		}
	})

	t.Run("case 2 run", func(t *testing.T) {
		tc := &Collector{}
		tc.Run("test1", func(t TB) {
			t.Logf("ok")
		})
		False(t, tc.Failed())
		tc.Run("test2", func(t TB) {
			t.Errorf("errorf")
		})
		True(t, tc.Failed())
		for _, line := range tc.getLogs() {
			t.Logf("line: %v", line)
			Contains(t, line.log, "xt.TestCollector")
			Contains(t, line.log, "testing_test.go")
		}
	})

	t.Run("case 3 dup name", func(t *testing.T) {
		tc := &Collector{}
		tc.Run("hello", func(t TB) {
			t.Logf("hello")
		})
		tc.Run("hello", func(t TB) {
			t.Logf("world")
		})
		False(t, tc.Failed())
		for _, line := range tc.getLogs() {
			t.Logf("line: %v", line)
			Contains(t, line.log, "xt.TestCollector")
			Contains(t, line.log, "testing_test.go")
		}
	})
}
