//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-18

package xsync

import (
	"sync"
	"sync/atomic"
)

func NewValue[T any](defaultValue T) *Value[T] {
	v := &Value[T]{}
	v.Store(defaultValue)
	return v
}

// Value 有默认值的 Value
type Value[T any] struct {
	value atomic.Value
	once  sync.Once
}

func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	v.once.Do(func() {
		v.value.Store(baggage[T]{})
	})
	return v.value.CompareAndSwap(baggage[T]{Value: old}, baggage[T]{Value: new})
}

func (v *Value[T]) Load() (val T) {
	v.once.Do(func() {
		v.value.Store(baggage[T]{})
	})
	value, ok := v.value.Load().(baggage[T])
	if ok {
		return value.Value
	}
	return val
}

func fnEmpty() {}

func (v *Value[T]) Store(val T) {
	v.once.Do(fnEmpty)
	v.value.Store(baggage[T]{Value: val})
}

func (v *Value[T]) Swap(new T) (old T) {
	v.once.Do(fnEmpty)
	value, ok := v.value.Swap(baggage[T]{Value: new}).(baggage[T])
	if ok {
		return value.Value
	}
	return old
}

// Clear 用空值覆盖
func (v *Value[T]) Clear() {
	var emp T
	v.Store(emp)
}

type baggage[T any] struct {
	Value T
}

// OnceLoadValue 只允许 load 一次的 Value
type OnceLoadValue[T any] struct {
	value atomic.Value
}

func (l *OnceLoadValue[T]) Store(val T) {
	l.value.Store(baggage[T]{Value: val})
}

func (l *OnceLoadValue[T]) Load() (val T) {
	old, ok := l.value.Swap(baggage[T]{}).(baggage[T])
	if ok {
		return old.Value
	}
	return val
}
