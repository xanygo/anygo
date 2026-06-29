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

// Value 泛型类型的 atomic.Value。
// 泛型类型 T 支持 interface：即可以存储同一个 interface 不同的实现。
type Value[T any] struct {
	value atomic.Value
	once  sync.Once
}

func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	v.once.Do(func() {
		v.value.Store(baggage1[T]{})
	})
	return v.value.CompareAndSwap(baggage1[T]{Value: old}, baggage1[T]{Value: new})
}

func (v *Value[T]) Load() (val T) {
	v.once.Do(func() {
		v.value.Store(baggage1[T]{})
	})
	value, _ := v.value.Load().(baggage1[T])
	return value.Value
}

func (v *Value[T]) Store(val T) {
	var set bool
	v.once.Do(func() {
		set = true
		v.value.Store(baggage1[T]{Value: val})
	})
	if !set {
		v.value.Store(baggage1[T]{Value: val})
	}
}

func (v *Value[T]) Swap(new T) (old T) {
	var set bool
	v.once.Do(func() {
		set = true
		v.value.Swap(baggage1[T]{Value: new})
	})
	if set {
		return old
	}
	value, _ := v.value.Swap(baggage1[T]{Value: new}).(baggage1[T])
	return value.Value
}

// Clear 用空值覆盖
func (v *Value[T]) Clear() {
	var emp T
	v.Store(emp)
}
