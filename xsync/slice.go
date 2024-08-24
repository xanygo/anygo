//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package xsync

import (
	"slices"
	"sync"
)

type Slice[T comparable] struct {
	items []T
	mux   sync.RWMutex
}

// Grow increases the slice's capacity, if necessary, to guarantee space for another n elements
func (s *Slice[T]) Grow(n int) {
	s.mux.Lock()
	s.items = slices.Grow(s.items, n)
	s.mux.Unlock()
}

func (s *Slice[T]) Append(v ...T) {
	s.mux.Lock()
	s.items = append(s.items, v...)
	s.mux.Unlock()
}

func (s *Slice[T]) Insert(index int, v ...T) {
	s.mux.Lock()
	s.items = slices.Insert(s.items, index, v...)
	s.mux.Unlock()
}

// Delete 删除 s[i:j] 之间的值
func (s *Slice[T]) Delete(i int, j int) {
	s.mux.Lock()
	s.items = slices.Delete(s.items, i, j)
	s.mux.Unlock()
}

// DeleteValue 删除指定的值
func (s *Slice[T]) DeleteValue(vs ...T) {
	if len(vs) == 0 {
		return
	}
	kv := make(map[T]struct{}, len(vs))
	for _, v := range vs {
		kv[v] = struct{}{}
	}
	s.mux.Lock()
	oldLen := len(s.items)
	for i := len(s.items) - 1; i >= 0; i-- {
		if _, ok := kv[s.items[i]]; ok {
			s.items = append(s.items[:i], s.items[i+1:]...)
		}
	}
	clear(s.items[len(s.items):oldLen])
	s.mux.Unlock()
}

// DeleteFunc removes any elements from s for which del returns true, returning the modified slice.
// DeleteFunc zeroes the elements between the new length and the original length.
func (s *Slice[T]) DeleteFunc(del func(T) bool) {
	s.mux.Lock()
	s.items = slices.DeleteFunc(s.items, del)
	s.mux.Unlock()
}

func (s *Slice[T]) Load() []T {
	s.mux.RLock()
	val := s.items
	s.mux.RUnlock()
	return val
}

func (s *Slice[T]) Clear() {
	s.mux.Lock()
	s.items = nil
	s.mux.Unlock()
}

func (s *Slice[T]) Store(all []T) {
	s.mux.Lock()
	s.items = slices.Clone(all)
	s.mux.Unlock()
}

func (s *Slice[T]) Swap(all []T) []T {
	s.mux.Lock()
	old := s.items
	s.items = slices.Clone(all)
	s.mux.Unlock()
	return old
}

func (s *Slice[T]) Head() (T, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.items) == 0 {
		var emp T
		return emp, false
	}
	return s.items[0], true
}

func (s *Slice[T]) Tail() (T, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	length := len(s.items)
	if length == 0 {
		var emp T
		return emp, false
	}
	return s.items[length-1], true
}

func (s *Slice[T]) Clone() *Slice[T] {
	return &Slice[T]{
		items: s.Load(),
	}
}
