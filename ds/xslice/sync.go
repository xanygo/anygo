//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

import (
	"iter"
	"slices"
	"sync"
)

// Sync 并发安全的 Slice
type Sync[T any] struct {
	items []T
	mux   sync.RWMutex
}

// Grow 分配容量
func (s *Sync[T]) Grow(n int) {
	s.mux.Lock()
	s.items = slices.Grow(s.items, n)
	s.mux.Unlock()
}

// Append 在尾部插入元素
func (s *Sync[T]) Append(v ...T) {
	s.mux.Lock()
	s.items = append(s.items, v...)
	s.mux.Unlock()
}

// Insert 在指定的位置插入元素
func (s *Sync[T]) Insert(index int, v ...T) {
	s.mux.Lock()
	s.items = slices.Insert(s.items, index, v...)
	s.mux.Unlock()
}

// Delete 删除 s[i:j] 之间的元素
func (s *Sync[T]) Delete(i int, j int) {
	s.mux.Lock()
	s.items = slices.Delete(s.items, i, j)
	s.mux.Unlock()
}

// DeleteFunc 删除满足回调函数的元素
func (s *Sync[T]) DeleteFunc(del func(T) bool) {
	s.mux.Lock()
	s.items = slices.DeleteFunc(s.items, del)
	s.mux.Unlock()
}

// Load 返回所有的值
func (s *Sync[T]) Load() []T {
	s.mux.RLock()
	val := s.items
	s.mux.RUnlock()
	return val
}

// Clear 清除所有值
func (s *Sync[T]) Clear() {
	s.mux.Lock()
	clear(s.items)
	s.items = nil
	s.mux.Unlock()
}

// Store 用传入的 slice 替换原有所有的值
func (s *Sync[T]) Store(all []T) {
	s.mux.Lock()
	s.items = all
	s.mux.Unlock()
}

// Swap 用传入的 slice 替换原有所有的值,并返回原有的值
func (s *Sync[T]) Swap(all []T) []T {
	s.mux.Lock()
	old := s.items
	s.items = all
	s.mux.Unlock()
	return old
}

// Head 读取头部的元素，若 slice 为空会返回 false
func (s *Sync[T]) Head() (T, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.items) == 0 {
		var emp T
		return emp, false
	}
	return s.items[0], true
}

// Tail 读取尾部的元素，若 slice 为空会返回 false
func (s *Sync[T]) Tail() (T, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	length := len(s.items)
	if length == 0 {
		var emp T
		return emp, false
	}
	return s.items[length-1], true
}

// PopHead 弹出头部的一个元素，若 slice 为空会返回 false
func (s *Sync[T]) PopHead() (val T, has bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items, val, has = PopHead(s.items)
	return val, has
}

// PopHeadN 弹出头部的 n 个元素，若 slice 为空会返回 false
func (s *Sync[T]) PopHeadN(n int) (values []T) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items, values = PopHeadN(s.items, n)
	return values
}

// PopTail 弹出尾部的一个元素，若 slice 为空会返回 false
func (s *Sync[T]) PopTail() (val T, has bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items, val, has = PopTail(s.items)
	return val, has
}

// PopTailN 弹出尾部的 n 个元素，若 slice 为空会返回 false
func (s *Sync[T]) PopTailN(n int) (values []T) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items, values = PopTailN(s.items, n)
	return values
}

func (s *Sync[T]) Clone() *Sync[T] {
	return &Sync[T]{
		items: slices.Clone(s.Load()),
	}
}

func (s *Sync[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.mux.RLock()
		defer s.mux.RUnlock()
		for _, val := range s.items {
			if !yield(val) {
				return
			}
		}
	}
}
