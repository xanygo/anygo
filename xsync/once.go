//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xsync

import (
	"sync"
)

type OnceDoErr = OnceDoValue[error]

type OnceDoValue[T any] struct {
	value T
	once  sync.Once
}

func (one *OnceDoValue[T]) Do(fn func() T) T {
	one.once.Do(func() {
		one.value = fn()
	})
	return one.value
}

type OnceDoValue2[M any, N any] struct {
	value1 M
	Value2 N
	once   sync.Once
}

func (one *OnceDoValue2[M, N]) Do(fn func() (M, N)) (M, N) {
	one.once.Do(func() {
		one.value1, one.Value2 = fn()
	})
	return one.value1, one.Value2
}

type OnceDoValueErr[T any] struct {
	value T
	err   error
	once  sync.Once
}

func (one *OnceDoValueErr[T]) Do(fn func() (T, error)) (T, error) {
	one.once.Do(func() {
		one.value, one.err = fn()
	})
	return one.value, one.err
}

type OnceDoValue3[A any, B any, C any] struct {
	value1 A
	Value2 B
	Value3 C
	once   sync.Once
}

func (one *OnceDoValue3[A, B, C]) Do(fn func() (A, B, C)) (A, B, C) {
	one.once.Do(func() {
		one.value1, one.Value2, one.Value3 = fn()
	})
	return one.value1, one.Value2, one.Value3
}

type OnceDoValue4[A any, B any, C any, D any] struct {
	value1 A
	Value2 B
	Value3 C
	Value4 D
	once   sync.Once
}

func (one *OnceDoValue4[A, B, C, D]) Do(fn func() (A, B, C, D)) (A, B, C, D) {
	one.once.Do(func() {
		one.value1, one.Value2, one.Value3, one.Value4 = fn()
	})
	return one.value1, one.Value2, one.Value3, one.Value4
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

func (os *OnceSet[T]) SetOnce(value T) {
	os.mux.RLock()
	has := os.has
	os.mux.RUnlock()
	if has {
		return
	}
	os.mux.Lock()
	if !os.has {
		os.has = true
		os.value = value
	}
	os.mux.Unlock()
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