//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package xsync

import "sync"

type Pool[T any] struct {
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() T

	sp   *sync.Pool
	once sync.Once
}

func (p *Pool[T]) init() {
	p.sp = &sync.Pool{
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
	return p.sp.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.onceInit()
	p.sp.Put(x)
}
