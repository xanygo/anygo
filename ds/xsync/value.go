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
	value, ok := v.value.Load().(baggage[T])
	if ok {
		return value.Value
	}
	var emp T
	return emp
}

func (v *Value[T]) Store(val T) {
	v.value.Store(baggage[T]{Value: val})
}

func (v *Value[T]) Swap(new T) (old T) {
	value, ok := v.value.Swap(baggage[T]{Value: new}).(baggage[T])
	if ok {
		return value.Value
	}
	var emp T
	return emp
}

type baggage[T any] struct {
	Value T
}
