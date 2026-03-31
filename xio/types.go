//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-31

package xio

import (
	"bufio"
	"time"
)

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
