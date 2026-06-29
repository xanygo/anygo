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

// Reset 重置 Done 的状态
func (o *Once) Reset() {
	o.done.Store(false)
}

type OnceDoErr = OnceDoValue[error]

type OnceDoValue[T any] struct {
	_     noCopy
	value Value[T]
	once  Once
}

func (one *OnceDoValue[T]) Do(fn func() T) T {
	one.once.Do(func() {
		one.value.Store(fn())
	})
	return one.value.Load()
}

// Done 返回 once.Get 是否已经执行的状态
func (one *OnceDoValue[T]) Done() bool {
	return one.once.Done()
}

// ResetDone 重置 Done 的状态
func (one *OnceDoValue[T]) ResetDone() {
	one.once.Reset()
}

func (one *OnceDoValue[T]) DoneValue() (ok bool, value T) {
	if one.once.Done() {
		return true, one.value.Load()
	}
	return false, value
}

type valuesAB[A any, B any] struct {
	A A
	B B
}

type OnceDoValue2[M any, N any] struct {
	_     noCopy
	value Value[*valuesAB[M, N]]
	once  Once
}

func (one *OnceDoValue2[M, N]) Do(fn func() (M, N)) (M, N) {
	one.once.Do(func() {
		a, b := fn()
		one.value.Store(&valuesAB[M, N]{A: a, B: b})
	})
	val := one.value.Load()
	return val.A, val.B
}

func (one *OnceDoValue2[M, N]) Done() bool {
	return one.once.Done()
}

// ResetDone 重置 Done 的状态
func (one *OnceDoValue2[M, N]) ResetDone() {
	one.once.Reset()
}

func (one *OnceDoValue2[M, N]) DoneValue() (ok bool, m M, n N) {
	if one.once.Done() {
		val := one.value.Load()
		return true, val.A, val.B
	}
	return false, m, n
}

type OnceDoValueErr[T any] struct {
	_     noCopy
	value Value[*valuesAB[T, error]]
	once  Once
}

func (one *OnceDoValueErr[T]) Do(fn func() (T, error)) (T, error) {
	one.once.Do(func() {
		a, b := fn()
		one.value.Store(&valuesAB[T, error]{A: a, B: b})
	})
	val := one.value.Load()
	return val.A, val.B
}

func (one *OnceDoValueErr[T]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValueErr[T]) ResetDone() {
	one.once.Reset()
}

// SetValue 设置值（并发安全，由于无锁，若和 Do 方法同时执行，不保证缓存是那个写入的）
func (one *OnceDoValueErr[T]) SetValue(v T, err error) {
	one.value.Store(&valuesAB[T, error]{A: v, B: err})
	one.once.Do(empty)
}

func (one *OnceDoValueErr[T]) DoneValue() (ok bool, v T, err error) {
	if one.once.Done() {
		val := one.value.Load()
		return true, val.A, val.B
	}
	return false, v, nil
}

func (one *OnceDoValueErr[T]) Value() (v T) {
	if one.once.Done() {
		return one.value.Load().A
	}
	return v
}

type valuesABC[A any, B any, C any] struct {
	A A
	B B
	C C
}

type OnceDoValue3[A any, B any, C any] struct {
	_     noCopy
	value Value[*valuesABC[A, B, C]]
	once  Once
}

func (one *OnceDoValue3[A, B, C]) Do(fn func() (A, B, C)) (A, B, C) {
	one.once.Do(func() {
		a, b, c := fn()
		one.value.Store(&valuesABC[A, B, C]{A: a, B: b, C: c})
	})
	val := one.value.Load()
	return val.A, val.B, val.C
}

func (one *OnceDoValue3[A, B, C]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValue3[A, B, C]) ResetDone() {
	one.once.Reset()
}

func (one *OnceDoValue3[A, B, C]) DoneValue() (ok bool, a A, b B, c C) {
	if one.once.Done() {
		val := one.value.Load()
		return true, val.A, val.B, val.C
	}
	return false, a, b, c
}

type valuesABCD[A any, B any, C any, D any] struct {
	A A
	B B
	C C
	D D
}

type OnceDoValue4[A any, B any, C any, D any] struct {
	_     noCopy
	value Value[*valuesABCD[A, B, C, D]]
	once  Once
}

func (one *OnceDoValue4[A, B, C, D]) Do(fn func() (A, B, C, D)) (A, B, C, D) {
	one.once.Do(func() {
		a, b, c, d := fn()
		one.value.Store(&valuesABCD[A, B, C, D]{A: a, B: b, C: c, D: d})
	})
	val := one.value.Load()
	return val.A, val.B, val.C, val.D
}

func (one *OnceDoValue4[A, B, C, D]) Done() bool {
	return one.once.Done()
}

func (one *OnceDoValue4[A, B, C, D]) ResetDone() {
	one.once.Reset()
}

func (one *OnceDoValue4[A, B, C, D]) DoneValue() (ok bool, a A, b B, c C, d D) {
	if one.once.Done() {
		val := one.value.Load()
		return true, val.A, val.B, val.C, val.D
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
	var ret bool
	os.mux.Lock()
	if !os.has {
		os.has = true
		os.value = value
		ret = true
	}
	os.mux.Unlock()
	return ret
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
	var set bool
	oi.once.Do(func() {
		set = true
		oi.value.Store(val)
	})
	if !set {
		oi.value.Store(val)
	}
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

type baggage1[T any] struct {
	Value T
}

type baggage2[T any] struct {
	Value T
	ok    bool
}

// OnceRead 只允许读取一次的 Value
type OnceRead[T any] struct {
	value atomic.Value
}

func (l *OnceRead[T]) Store(val T) {
	l.value.Store(baggage2[T]{Value: val, ok: true})
}

// Read 在 Store 之后，第一次读取，返回 <值,true>；之后返回 <空值,false>。
// 若之前没有 Store，则返回 <空值,false>。
func (l *OnceRead[T]) Read() (val T, ok bool) {
	old, ok := l.value.Swap(baggage2[T]{}).(baggage2[T])
	if ok && old.ok {
		return old.Value, true
	}
	return val, false
}
