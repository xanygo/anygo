//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-29

package xsync

import (
	"bytes"
	"io"
	"sync"
)

func NewBytesBuffer(buf []byte) *BytesBuffer {
	b := &BytesBuffer{}
	b.Write(buf)
	return b
}

func NewBytesBufferString(s string) *BytesBuffer {
	b := &BytesBuffer{}
	b.WriteString(s)
	return b
}

// BytesBuffer 使用 sync.Mutex 封装的 bytes.Buffer
type BytesBuffer struct {
	bf  bytes.Buffer
	mux sync.Mutex
}

func (b *BytesBuffer) WithMutex(fn func(bf *bytes.Buffer)) {
	b.mux.Lock()
	defer b.mux.Unlock()
	fn(&b.bf)
}

func (b *BytesBuffer) Bytes() []byte {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.Bytes()
}

func (b *BytesBuffer) AvailableBuffer() []byte {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.AvailableBuffer()
}

func (b *BytesBuffer) String() string {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.String()
}

func (b *BytesBuffer) Len() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.Len()
}

func (b *BytesBuffer) Cap() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.Cap()
}

func (b *BytesBuffer) Truncate(n int) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.bf.Truncate(n)
}

func (b *BytesBuffer) Reset() {
	b.mux.Lock()
	b.bf.Reset()
	b.mux.Unlock()
}

func (b *BytesBuffer) Grow(n int) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.bf.Grow(n)
}

func (b *BytesBuffer) Write(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.Write(p)
}

func (b *BytesBuffer) WriteString(s string) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.WriteString(s)
}

func (b *BytesBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.ReadFrom(r)
}

func (b *BytesBuffer) WriteTo(w io.Writer) (n int64, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.WriteTo(w)
}

func (b *BytesBuffer) WriteByte(c byte) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.WriteByte(c)
}

func (b *BytesBuffer) WriteRune(r rune) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.WriteRune(r)
}

func (b *BytesBuffer) Read(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.Read(p)
}

func (b *BytesBuffer) Next(n int) []byte {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.Next(n)
}

func (b *BytesBuffer) ReadByte() (byte, error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.ReadByte()
}

func (b *BytesBuffer) ReadRune() (r rune, size int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.ReadRune()
}

func (b *BytesBuffer) UnreadRune() error {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.UnreadRune()
}

func (b *BytesBuffer) UnreadByte() error {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.UnreadByte()
}

func (b *BytesBuffer) ReadBytes(delim byte) (line []byte, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.ReadBytes(delim)
}

func (b *BytesBuffer) ReadString(delim byte) (line string, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.bf.ReadString(delim)
}

func (b *BytesBuffer) Unwrap() *bytes.Buffer {
	return &b.bf
}
