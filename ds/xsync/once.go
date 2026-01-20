//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xsync

import (
	"context"
	"sync"
	"sync/atomic"
)

type Once struct {
	_    noCopy
	done atomic.Bool
	m    sync.Mutex
}

func (o *Once) Do(f func()) {
	if !o.done.Load() {
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.done.Load() {
		defer o.done.Store(true)
		f()
	}
}

// Done 返回是否已经执行的状态
func (o *Once) Done() bool {
	return o.done.Load()
}

type OnceDoErr = OnceDoValue[error]

type OnceDoValue[T any] struct {
	_     noCopy
	value T
	once  Once
}

func (one *OnceDoValue[T]) Do(fn func() T) T {
	one.once.Do(func() {
		one.value = fn()
	})
	return one.value
}

func (one *OnceDoValue[T]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValue[T]) DoneValue() (ok bool, value T) {
	if one.once.Done() {
		return true, one.value
	}
	return false, value
}

type OnceDoValue2[M any, N any] struct {
	_      noCopy
	value1 M
	Value2 N
	once   Once
}

func (one *OnceDoValue2[M, N]) Do(fn func() (M, N)) (M, N) {
	one.once.Do(func() {
		one.value1, one.Value2 = fn()
	})
	return one.value1, one.Value2
}

func (one *OnceDoValue2[M, N]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValue2[M, N]) DoneValue() (ok bool, m M, n N) {
	if one.once.Done() {
		return true, one.value1, one.Value2
	}
	return false, m, n
}

type OnceDoValueErr[T any] struct {
	_     noCopy
	value T
	err   error
	once  Once
}

func (one *OnceDoValueErr[T]) Do(fn func() (T, error)) (T, error) {
	one.once.Do(func() {
		one.value, one.err = fn()
	})
	return one.value, one.err
}

func (one *OnceDoValueErr[T]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValueErr[T]) DoneValue() (ok bool, v T, err error) {
	if one.once.Done() {
		return true, one.value, one.err
	}
	return false, v, nil
}

type OnceDoValue3[A any, B any, C any] struct {
	_      noCopy
	value1 A
	Value2 B
	Value3 C
	once   Once
}

func (one *OnceDoValue3[A, B, C]) Do(fn func() (A, B, C)) (A, B, C) {
	one.once.Do(func() {
		one.value1, one.Value2, one.Value3 = fn()
	})
	return one.value1, one.Value2, one.Value3
}

func (one *OnceDoValue3[A, B, C]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValue3[A, B, C]) DoneValue() (ok bool, a A, b B, c C) {
	if one.once.Done() {
		return true, one.value1, one.Value2, one.Value3
	}
	return false, a, b, c
}

type OnceDoValue4[A any, B any, C any, D any] struct {
	_      noCopy
	value1 A
	Value2 B
	Value3 C
	Value4 D
	once   Once
}

func (one *OnceDoValue4[A, B, C, D]) Do(fn func() (A, B, C, D)) (A, B, C, D) {
	one.once.Do(func() {
		one.value1, one.Value2, one.Value3, one.Value4 = fn()
	})
	return one.value1, one.Value2, one.Value3, one.Value4
}

func (one *OnceDoValue4[A, B, C, D]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValue4[A, B, C, D]) DoneValue() (ok bool, a A, b B, c C, d D) {
	if one.once.Done() {
		return true, one.value1, one.Value2, one.Value3, one.Value4
	}
	return false, a, b, c, d
}

func OnceValue[T any](fn func() T) func() T {
	return sync.OnceValue[T](fn)
}

func OnceValue2[A any, B any](fn func() (A, B)) func() (A, B) {
	var once OnceDoValue2[A, B]
	return func() (A, B) {
		return once.Do(fn)
	}
}

func OnceValue3[A any, B any, C any](fn func() (A, B, C)) func() (A, B, C) {
	var once OnceDoValue3[A, B, C]
	return func() (A, B, C) {
		return once.Do(fn)
	}
}

func OnceValue4[A any, B any, C any, D any](fn func() (A, B, C, D)) func() (A, B, C, D) {
	var once OnceDoValue4[A, B, C, D]
	return func() (A, B, C, D) {
		return once.Do(fn)
	}
}

// OnceSet can Set Value only Once
type OnceSet[T any] struct {
	value T
	has   bool
	mux   sync.RWMutex
}

func (os *OnceSet[T]) SetOnce(value T) bool {
	os.mux.RLock()
	has := os.has
	os.mux.RUnlock()
	if has {
		return false
	}
	os.mux.Lock()
	if !os.has {
		os.has = true
		os.value = value
	}
	os.mux.Unlock()
	return true
}

func (os *OnceSet[T]) Get() (T, bool) {
	os.mux.RLock()
	defer os.mux.RUnlock()
	return os.value, os.has
}

func (os *OnceSet[T]) Replace(value T) {
	os.mux.Lock()
	os.value = value
	os.has = true
	os.mux.Unlock()
}

func (os *OnceSet[T]) Clear() {
	os.mux.Lock()
	var emp T
	os.value = emp
	os.has = false
	os.mux.Unlock()
}

// OnceInit 延迟初始化一次
type OnceInit[T any] struct {
	// New 必填，在初始化的时候调用一次
	New func() T

	once  sync.Once
	value Value[T]
}

func (oi *OnceInit[T]) doInit() {
	oi.value.Store(oi.New())
}

func (oi *OnceInit[T]) Load() T {
	oi.once.Do(oi.doInit)
	return oi.value.Load()
}

func (oi *OnceInit[T]) Store(val T) {
	oi.once.Do(empty)
	oi.value.Store(val)
}

func (oi *OnceInit[T]) Swap(new T) (old T) {
	oi.once.Do(empty)
	return oi.value.Swap(new)
}

func (oi *OnceInit[T]) CompareAndSwap(old, new T) (swapped bool) {
	oi.once.Do(empty)
	return oi.value.CompareAndSwap(old, new)
}

func empty() {}

// OnceInitCtx 延迟初始化一次
type OnceInitCtx[T any] struct {
	// New 必填，在初始化的时候调用一次
	New func(ctx context.Context) T

	once  sync.Once
	value Value[T]
}

func (oi *OnceInitCtx[T]) InitOnce(ctx context.Context) {
	oi.once.Do(func() {
		value := oi.New(ctx)
		oi.value.Store(value)
	})
}

func (oi *OnceInitCtx[T]) Load() T {
	return oi.value.Load()
}

func (oi *OnceInitCtx[T]) Store(val T) {
	oi.value.Store(val)
}

func (oi *OnceInitCtx[T]) Swap(new T) (old T) {
	return oi.value.Swap(new)
}

func (oi *OnceInitCtx[T]) CompareAndSwap(old, new T) (swapped bool) {
	return oi.value.CompareAndSwap(old, new)
}
