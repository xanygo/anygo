//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package xslice

import (
	"fmt"
	"sync"
)

// NewRing 创建新的 Ring，caption-容量，应 > 0
func NewRing[T any](caption int) *Ring[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid Ring caption %d", caption))
	}
	return &Ring[T]{
		caption: caption,
		values:  make([]T, caption),
		mux:     &sync.RWMutex{},
	}
}

// Ring 具有指定最大容量的，环形结构的 slice，容量满的情况下，新元素会覆盖老元素
type Ring[T any] struct {
	values  []T
	caption int
	length  int
	index   int
	mux     *sync.RWMutex
}

// Add 添加新的元素，容量满的情况下，会覆盖老的值
func (r *Ring[T]) Add(values ...T) {
	if len(values) == 0 {
		return
	}
	r.mux.Lock()
	for _, v := range values {
		r.values[r.index] = v
		r.index++
		if r.index == r.caption {
			r.index = 0
		}
		if r.length < r.caption {
			r.length++
		}
	}
	r.mux.Unlock()
}

// AddSwap 添加并返回被替换的值
func (r *Ring[T]) AddSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}
	r.mux.Unlock()
	return old, swapped
}

func (r *Ring[T]) Len() int {
	r.mux.RLock()
	val := r.length
	r.mux.RUnlock()
	return val
}

// Range 遍历，先加入的会先遍历
func (r *Ring[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

// Values 返回所有值，先加入的排在前面
func (r *Ring[T]) Values() []T {
	r.mux.RLock()
	defer r.mux.RUnlock()
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func NewRingUnique[T comparable](caption int) *RingUnique[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid Ring caption %d", caption))
	}
	return &RingUnique[T]{
		caption:    caption,
		values:     make([]T, caption),
		valueIndex: make(map[T]int, caption),
		mux:        &sync.RWMutex{},
	}
}

// RingUnique 具有唯一值的 ring list
type RingUnique[T comparable] struct {
	values     []T
	valueIndex map[T]int
	caption    int
	length     int
	index      int
	mux        *sync.RWMutex
}

// Add 添加新的元素，容量满的情况下，会覆盖老的值
func (r *RingUnique[T]) Add(values ...T) {
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, v := range values {
		oldIndex, has := r.valueIndex[v]
		if has {
			r.values[oldIndex] = v
			continue
		}

		r.values[r.index] = v
		r.valueIndex[v] = r.index
		r.index++
		if r.index == r.caption {
			r.index = 0
		}
		if r.length < r.caption {
			r.length++
		}
	}
}

// AddSwap 添加并返回被替换的值
func (r *RingUnique[T]) AddSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	oldIndex, has := r.valueIndex[v]
	if has {
		old = r.values[oldIndex]
		r.values[oldIndex] = v
		return old, true
	}

	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.valueIndex[v] = r.index
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}

	return old, swapped
}

func (r *RingUnique[T]) Len() int {
	r.mux.RLock()
	val := r.length
	r.mux.RUnlock()
	return val
}

// Range 遍历，先加入的会先遍历
func (r *RingUnique[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

// Values 返回所有值，先加入的排在前面
func (r *RingUnique[T]) Values() []T {
	r.mux.RLock()
	defer r.mux.RUnlock()
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}
