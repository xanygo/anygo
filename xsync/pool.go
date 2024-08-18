//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package xsync

import (
	"bytes"
	"sync"
)

type Pool[T any] struct {
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() T

	pool *sync.Pool
	once sync.Once
}

func (p *Pool[T]) init() {
	p.pool = &sync.Pool{
		New: func() any {
			return p.New()
		},
	}
}

func (p *Pool[T]) onceInit() {
	p.once.Do(p.init)
}

func (p *Pool[T]) Get() T {
	p.onceInit()
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.onceInit()
	p.pool.Put(x)
}

func NewBytesBufferPool(maxCap int) *BytesBufferPool {
	return &BytesBufferPool{
		MaxCap: maxCap,
	}
}

type BytesBufferPool struct {
	MaxCap int
	pool   *sync.Pool
	once   sync.Once
}

func (p *BytesBufferPool) initOnce() {
	p.once.Do(func() {
		p.pool = &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}
	})
}

func (p *BytesBufferPool) Get() *bytes.Buffer {
	p.initOnce()
	return p.pool.Get().(*bytes.Buffer)
}

func (p *BytesBufferPool) Put(bf *bytes.Buffer) {
	p.initOnce()
	if p.MaxCap > 0 && bf.Cap() > p.MaxCap {
		return
	}
	p.pool.Put(bf)
}

var DefaultBytesBufferPool = NewBytesBufferPool(1 << 20)

func GetBytesBuffer() *bytes.Buffer {
	return DefaultBytesBufferPool.Get()
}

func PutBytesBuffer(bf *bytes.Buffer) {
	DefaultBytesBufferPool.Put(bf)
}
