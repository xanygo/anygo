//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package xsync

import (
	"bytes"
	"sync"
)

func NewPool[T any](new func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return new()
			},
		},
	}
}

// Pool sync.Pool 的泛型封装
type Pool[T any] struct {
	pool *sync.Pool
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}

func NewBytesBufferPool(maxCap int) *BytesBufferPool {
	return &BytesBufferPool{
		maxCap: maxCap,
		pool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

// BytesBufferPool  BytesBuffer 的对象池，
// 若 buffer 的 caption > maxCap 时，该对象会被丢弃，以避免占用过多内存
type BytesBufferPool struct {
	maxCap int
	pool   *sync.Pool
}

func (p *BytesBufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *BytesBufferPool) Put(bf *bytes.Buffer) {
	bf.Reset()
	if p.maxCap > 0 && bf.Cap() > p.maxCap {
		return
	}
	p.pool.Put(bf)
}

// DefaultBytesBufferPool 全局的 BytesBuffer 对象池
var DefaultBytesBufferPool = NewBytesBufferPool(1 << 20)

// GetBytesBuffer 从全局 BytesBuffer 对象池获取一个新的 Buffer 对象
func GetBytesBuffer() *bytes.Buffer {
	return DefaultBytesBufferPool.Get()
}

// PutBytesBuffer 将 Buffer 对象放回全局的 BytesBuffer 对象池
func PutBytesBuffer(bf *bytes.Buffer) {
	DefaultBytesBufferPool.Put(bf)
}
