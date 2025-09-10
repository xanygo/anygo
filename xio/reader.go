//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xio

import "io"

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
