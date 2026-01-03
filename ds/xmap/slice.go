//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xmap

import (
	"iter"
	"maps"
	"sync"

	"github.com/xanygo/anygo/internal/zslice"
)

// SliceValue 值为 slice 的 非并发安全的 map ( map[K][]V )
type SliceValue[K, V comparable] struct {
	data map[K][]V
}

func (s *SliceValue[K, V]) Set(key K, values ...V) {
	if s.data == nil {
		s.data = make(map[K][]V)
	}
	s.data[key] = values
}

func (s *SliceValue[K, V]) Get(key K) []V {
	if len(s.data) == 0 {
		return nil
	}
	return s.data[key]
}

func (s *SliceValue[K, V]) GetFirst(key K) (v V) {
	vs := s.Get(key)
	if len(vs) == 0 {
		return v
	}
	return vs[0]
}

func (s *SliceValue[K, V]) AddUnique(key K, values ...V) {
	if len(values) == 0 {
		return
	}
	if s.data == nil {
		s.data = make(map[K][]V)
	}
	s.data[key] = append(s.data[key], values...)
	zslice.Unique(s.data[key])
}

func (s *SliceValue[K, V]) Delete(keys ...K) {
	if len(s.data) == 0 {
		return
	}
	for _, key := range keys {
		delete(s.data, key)
	}
}

func (s *SliceValue[K, V]) DeleteValue(key K, values ...V) {
	if len(s.data) == 0 {
		return
	}
	vs, ok := s.data[key]
	if !ok {
		return
	}
	vs = zslice.DeleteValue(vs, values...)
	if len(vs) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = vs
	}
}

func (s *SliceValue[K, V]) Has(key K) bool {
	if len(s.data) == 0 {
		return false
	}

	_, has := s.data[key]
	return has
}

func (s *SliceValue[K, V]) HasValue(key K, values ...V) bool {
	if len(values) == 0 {
		return false
	}
	if len(s.data) == 0 {
		return false
	}

	return zslice.ContainsAny(s.data[key], values...)
}

func (s *SliceValue[K, V]) Keys() []K {
	return Keys(s.data)
}

func (s *SliceValue[K, V]) Map(clone bool) map[K][]V {
	if !clone {
		return s.data
	}
	return maps.Clone(s.data)
}

func (s *SliceValue[K, V]) Iter() iter.Seq2[K, []V] {
	return func(yield func(K, []V) bool) {
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *SliceValue[K, V]) Len() int {
	return len(s.data)
}

// -----

// SliceValueSync 值为 slice 的 并发安全的 map ( map[K][]V )
type SliceValueSync[K, V comparable] struct {
	data map[K][]V
	mux  sync.RWMutex
}

func (s *SliceValueSync[K, V]) Set(key K, values ...V) {
	s.mux.Lock()
	if s.data == nil {
		s.data = make(map[K][]V)
	}
	s.data[key] = values
	s.mux.Unlock()
}

func (s *SliceValueSync[K, V]) Get(key K) []V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 {
		return nil
	}
	return s.data[key]
}

func (s *SliceValueSync[K, V]) GetFirst(key K) (v V) {
	vs := s.Get(key)
	if len(vs) == 0 {
		return v
	}
	return vs[0]
}

func (s *SliceValueSync[K, V]) AddUnique(key K, values ...V) {
	if len(values) == 0 {
		return
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.data == nil {
		s.data = make(map[K][]V)
	}
	s.data[key] = append(s.data[key], values...)
	zslice.Unique(s.data[key])
}

func (s *SliceValueSync[K, V]) Delete(keys ...K) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if len(s.data) == 0 {
		return
	}
	for _, key := range keys {
		delete(s.data, key)
	}
}

func (s *SliceValueSync[K, V]) DeleteValue(key K, values ...V) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if len(s.data) == 0 {
		return
	}
	vs, ok := s.data[key]
	if !ok {
		return
	}
	vs = zslice.DeleteValue(vs, values...)
	if len(vs) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = vs
	}
}

func (s *SliceValueSync[K, V]) Has(key K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 {
		return false
	}

	_, has := s.data[key]
	return has
}

func (s *SliceValueSync[K, V]) HasValue(key K, values ...V) bool {
	if len(values) == 0 {
		return false
	}
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 || len(s.data[key]) == 0 {
		return false
	}

	return zslice.ContainsAny(s.data[key], values...)
}

func (s *SliceValueSync[K, V]) Keys() []K {
	s.mux.RLock()
	result := Keys(s.data)
	s.mux.RUnlock()
	return result
}

func (s *SliceValueSync[K, V]) Map(clone bool) (result map[K][]V) {
	s.mux.RLock()
	if clone {
		result = maps.Clone(s.data)
	} else {
		result = s.data
	}
	s.mux.RUnlock()
	return result
}

// Iter 用于遍历，在遍历期间会锁住整个对象
func (s *SliceValueSync[K, V]) Iter() iter.Seq2[K, []V] {
	return func(yield func(K, []V) bool) {
		s.mux.RLock()
		defer s.mux.RUnlock()
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Rand 随机返回一个
func (s *SliceValueSync[K, V]) Rand() (key K, value []V, ok bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for k, v := range s.data {
		return k, v, true
	}
	return key, nil, false
}

func (s *SliceValueSync[K, V]) Len() int {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.data)
}
