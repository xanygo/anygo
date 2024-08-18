//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-18

package xio

import (
	"io"
	"net/http"
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
