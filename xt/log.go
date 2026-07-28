package xt

import (
	"io"
	"os"
	"sync"
)

type TLogf interface {
	Logf(format string, args ...any)
}

var _ io.Writer = (*TLogWriter)(nil)

// TLogWriter 用于单侧场景，可以结合 interceptor 将日志输出到 *testing.Logf
type TLogWriter struct {
	w   TLogf
	mux sync.Mutex

	Default io.Writer
}

func (t *TLogWriter) Switch(w TLogf) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.w = w
}

func (t *TLogWriter) Write(p []byte) (n int, err error) {
	t.mux.Lock()
	w := t.w
	t.mux.Unlock()
	if w == nil {
		if t.Default != nil {
			return t.Default.Write(p)
		}
		return os.Stderr.Write(p)
	}
	w.Logf(string(p))
	return len(p), nil
}
