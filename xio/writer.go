//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-18

package xio

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Flusher 具有 Flush 方法的接口定义
type Flusher interface {
	Flush() error
}

// TryFlush 尝试调用 writer的 Flush 方法，若不支持会直接返回 nil
func TryFlush(w io.Writer) error {
	switch fw := w.(type) {
	case Flusher:
		return fw.Flush()
	case http.Flusher:
		fw.Flush()
		return nil
	default:
		return nil
	}
}

// ResetWriter 支持重新设置 Writer 的接口定义
type ResetWriter interface {
	io.Writer
	Reset(w io.Writer)
}

// NewResetWriter 封装一个 writer 为 ResetWriter
func NewResetWriter(w io.Writer) ResetWriter {
	if rw, ok := w.(*resetWriter); ok {
		return rw
	}
	return &resetWriter{
		raw: w,
	}
}

type resetWriter struct {
	raw io.Writer
	mux sync.RWMutex
}

func (w *resetWriter) Write(p []byte) (n int, err error) {
	w.mux.RLock()
	raw := w.raw
	w.mux.RUnlock()
	return raw.Write(p)
}

func (w *resetWriter) Reset(raw io.Writer) {
	w.mux.Lock()
	w.raw = raw
	w.mux.Unlock()
}

type FlushWriter interface {
	Flusher
	io.Writer
}

// NopWriteCloser 将一个 writer 封装为具有 空 Close 方法的 WriteCloser
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{Writer: w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

type StringWriter interface {
	WriteString(s string) (int, error)
}

func WriteStrings(bf StringWriter, ss ...string) (int, error) {
	var total int
	for _, str := range ss {
		n, err := bf.WriteString(str)
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

type PrintfWriter interface {
	Printf(string, ...any)
}

func AsPrintfWriter(w io.Writer) PrintfWriter {
	return &pw{w: w}
}

var _ PrintfWriter = (*pw)(nil)

type pw struct {
	w io.Writer
}

func (p *pw) Printf(str string, args ...any) {
	_, _ = fmt.Fprintf(p.w, str, args...)
}
