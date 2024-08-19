//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-18

package xio

import (
	"io"
	"net/http"
	"sync"
)

// Flusher can flush
type Flusher interface {
	Flush() error
}

// TryFlush try flush
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

// ResetWriter writer can reset
type ResetWriter interface {
	io.Writer
	Reset(w io.Writer)
}

// NewResetWriter wrap writer to ResetWriter
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

// NopWriteCloser nop closer for writer
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{Writer: w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}
