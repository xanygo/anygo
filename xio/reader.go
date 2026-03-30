//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xio

import (
	"bufio"
	"io"
	"time"
)

func LimitReaderCloser(rd io.ReadCloser, size int64) io.ReadCloser {
	return &limitReadCloser{
		raw:     rd,
		limiter: io.LimitReader(rd, size),
	}
}

type limitReadCloser struct {
	raw     io.ReadCloser
	limiter io.Reader
}

func (l *limitReadCloser) Read(p []byte) (n int, err error) {
	return l.limiter.Read(p)
}

func (l *limitReadCloser) Close() error {
	return l.raw.Close()
}

var _ StringReader = (*bufio.Reader)(nil)

type StringReader interface {
	ReadString(delim byte) (string, error)
}

var _ SliceReader = (*bufio.Reader)(nil)

type SliceReader interface {
	ReadSlice(delim byte) (line []byte, err error)
}

var _ BytesReader = (*bufio.Reader)(nil)

type BytesReader interface {
	ReadBytes(delim byte) ([]byte, error)
}

type DeadlineSetter interface {
	SetDeadline(t time.Time) error
}

type ReadDeadlineSetter interface {
	SetReadDeadline(t time.Time) error
}

type WriteDeadlineSetter interface {
	SetWriteDeadline(t time.Time) error
}
