//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xmap

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"sync"
)

// Ordered 按照写入顺序排序的 Sync, 非并发安全的
type Ordered[K comparable, V any] struct {
	// Caption 初始化 map 时，默认的容量，可选，默认值为 4
	Caption int

	keys []K
	db   map[K]V
}

func (s *Ordered[K, V]) Set(key K, value V) {
	if s.db == nil {
		s.db = make(map[K]V, max(4, s.Caption))
	}
	_, has := s.db[key]
	s.db[key] = value
	if !has {
		s.keys = append(s.keys, key)
	}
}

func (s *Ordered[K, V]) Delete(keys ...K) {
	if s.db == nil {
		return
	}
	for _, key := range keys {
		_, has := s.db[key]
		if !has {
			continue
		}
		delete(s.db, key)
		index := slices.Index(s.keys, key)
		s.keys = slices.Delete(s.keys, index, index+1)
	}
}

func (s *Ordered[K, V]) Get(key K) (V, bool) {
	if s.db == nil {
		var emp V
		return emp, false
	}
	v, ok := s.db[key]
	return v, ok
}

func (s *Ordered[K, V]) GetDf(key K, def V) V {
	if s.db == nil {
		return def
	}
	v, ok := s.db[key]
	if ok {
		return v
	}
	return def
}

func (s *Ordered[K, V]) MustGet(key K) V {
	if s.db != nil {
		if v, ok := s.db[key]; ok {
			return v
		}
	}
	panic(fmt.Sprintf("not found key=%v", key))
}

func (s *Ordered[K, V]) Has(key K) bool {
	if s.db == nil {
		return false
	}
	_, ok := s.db[key]
	return ok
}

func (s *Ordered[K, V]) HasAny(keys ...K) bool {
	if s.db == nil {
		return false
	}
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			return true
		}
	}
	return false
}

func (s *Ordered[K, V]) Range(fn func(key K, value V) bool) {
	for _, k := range s.keys {
		if !fn(k, s.db[k]) {
			return
		}
	}
}

func (s *Ordered[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		s.Range(yield)
	}
}

func (s *Ordered[K, V]) Keys() []K {
	return s.keys
}

func (s *Ordered[K, V]) Map(clone bool) map[K]V {
	if clone {
		return maps.Clone(s.db)
	}
	return s.db
}

func (s *Ordered[K, V]) KVs(clone bool) ([]K, map[K]V) {
	if clone {
		return slices.Clone(s.keys), maps.Clone(s.db)
	}
	return s.keys, s.db
}

func (s *Ordered[K, V]) Values() []V {
	if s.db == nil {
		return nil
	}
	values := make([]V, len(s.keys))
	for index, key := range s.keys {
		values[index] = s.db[key]
	}
	return values
}

func (s *Ordered[K, V]) Len() int {
	return len(s.keys)
}

func (s *Ordered[K, V]) Clear() {
	s.keys = nil
	clear(s.db)
}

func (s *Ordered[K, V]) Clone() *Ordered[K, V] {
	keys, values := s.KVs(true)
	return &Ordered[K, V]{
		keys: keys,
		db:   values,
	}
}

// ----------------------------------

// OrderedSync 按照写入顺序排序的 Sync, 并发安全的
type OrderedSync[K comparable, V any] struct {
	// Caption 初始化 map 时，默认的容量，可选，默认值为 4
	Caption int

	keys []K
	db   map[K]V
	mux  sync.RWMutex
}

func (s *OrderedSync[K, V]) Set(key K, value V) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.db == nil {
		s.db = make(map[K]V, max(4, s.Caption))
	}
	_, has := s.db[key]
	s.db[key] = value
	if !has {
		s.keys = append(s.keys, key)
	}
}

func (s *OrderedSync[K, V]) Delete(keys ...K) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if len(s.db) == 0 {
		return
	}
	for _, key := range keys {
		_, has := s.db[key]
		if !has {
			continue
		}
		delete(s.db, key)
		index := slices.Index(s.keys, key)
		s.keys = slices.Delete(s.keys, index, index+1)
	}
}

func (s *OrderedSync[K, V]) Get(key K) (V, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.db) == 0 {
		var emp V
		return emp, false
	}
	v, ok := s.db[key]
	return v, ok
}

func (s *OrderedSync[K, V]) GetDf(key K, def V) V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.db) == 0 {
		return def
	}
	v, ok := s.db[key]
	if ok {
		return v
	}
	return def
}

func (s *OrderedSync[K, V]) MustGet(key K) V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.db) != 0 {
		if v, ok := s.db[key]; ok {
			return v
		}
	}
	panic(fmt.Sprintf("not found key=%v", key))
}

func (s *OrderedSync[K, V]) Has(key K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.db) == 0 {
		return false
	}
	_, ok := s.db[key]
	return ok
}

func (s *OrderedSync[K, V]) HasAny(keys ...K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.db) == 0 {
		return false
	}
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			return true
		}
	}
	return false
}

func (s *OrderedSync[K, V]) Range(fn func(key K, value V) bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, k := range s.keys {
		if !fn(k, s.db[k]) {
			return
		}
	}
}

func (s *OrderedSync[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		s.Range(yield)
	}
}

func (s *OrderedSync[K, V]) Keys() []K {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.keys
}

func (s *OrderedSync[K, V]) Map(clone bool) map[K]V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if clone {
		return maps.Clone(s.db)
	}
	return s.db
}

func (s *OrderedSync[K, V]) KVs(clone bool) ([]K, map[K]V) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if clone {
		return slices.Clone(s.keys), maps.Clone(s.db)
	}
	return s.keys, s.db
}

func (s *OrderedSync[K, V]) Values() []V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.db) == 0 {
		return nil
	}
	values := make([]V, len(s.keys))
	for index, key := range s.keys {
		values[index] = s.db[key]
	}
	return values
}

func (s *OrderedSync[K, V]) Len() int {
	s.mux.RLock()
	val := len(s.keys)
	s.mux.RUnlock()
	return val
}

func (s *OrderedSync[K, V]) Clear() {
	s.mux.Lock()
	s.keys = nil
	clear(s.db)
	s.mux.Unlock()
}

func (s *OrderedSync[K, V]) Clone() *OrderedSync[K, V] {
	keys, values := s.KVs(true)
	return &OrderedSync[K, V]{
		keys: keys,
		db:   values,
	}
}
