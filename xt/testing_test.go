// Copyright(C) 2024 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123+git@gmail.com>
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
	wantSuccess bool // 期望成功
	statusOk    bool // 运行状态是否正常
	lastCaller  string
}

func (m *myTesting) Helper() {
	_, file, lineNo, _ := runtime.Caller(1)
	m.lastCaller = fmt.Sprintf("%s:%d", filepath.Base(file), lineNo)
}

func (m *myTesting) Fatalf(format string, args ...any) {
	m.statusOk = false
	m.t.Helper()
	if m.wantSuccess {
		msg := fmt.Sprintf(format, args...)
		m.t.Fatalf("%s: expect success, but not, "+msg, m.lastCaller)
	}
}

// Success 期望传入的 fn 都运行成功
// fn 里一次能写多条断言代码
func (m *myTesting) Success(fns ...func(t Testing)) {
	m.t.Run("success", func(t *testing.T) {
		for i, fn := range fns {
			t.Run(fmt.Sprintf("fn-%d", i), func(t *testing.T) {
				m1 := &myTesting{
					t:           t,
					wantSuccess: true,
					statusOk:    true,
				}
				fn(m1)
			})
		}
	})
}

// Fail 期望传入的 fn 都运行失败
// fn 里一次只能写一条断言代码
func (m *myTesting) Fail(fns ...func(t Testing)) {
	m.t.Run("Fail", func(t *testing.T) {
		for i, fn := range fns {
			t.Run(fmt.Sprintf("fn-%d", i), func(t *testing.T) {
				m1 := &myTesting{
					t:           t,
					wantSuccess: false,
					statusOk:    true,
				}
				fn(m1)
				if m1.statusOk {
					m1.t.Fatalf("%s: expect fail but not", m.lastCaller)
				}
			})
		}
	})
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
