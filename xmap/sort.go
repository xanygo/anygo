//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xmap

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

// Sorted 按照写入顺序排序的 Value
type Sorted[K comparable, V any] struct {
	keys []K
	db   map[K]V
	mux  sync.RWMutex
}

func (s *Sorted[K, V]) Set(key K, value V) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.db == nil {
		s.db = make(map[K]V, 8)
	}
	_, has := s.db[key]
	s.db[key] = value
	if !has {
		s.keys = append(s.keys, key)
	}
}

func (s *Sorted[K, V]) Delete(keys ...K) {
	s.mux.Lock()
	defer s.mux.Unlock()
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

func (s *Sorted[K, V]) Get(key K) (V, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.db == nil {
		var emp V
		return emp, false
	}
	v, ok := s.db[key]
	return v, ok
}

func (s *Sorted[K, V]) GetDefault(key K, def V) V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.db == nil {
		return def
	}
	v, ok := s.db[key]
	if ok {
		return v
	}
	return def
}

func (s *Sorted[K, V]) MustGet(key K) V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.db != nil {
		if v, ok := s.db[key]; ok {
			return v
		}
	}
	panic(fmt.Sprintf("not found key=%v", key))
}

func (s *Sorted[K, V]) Has(key K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if s.db == nil {
		return false
	}
	_, ok := s.db[key]
	return ok
}
func (s *Sorted[K, V]) HasAny(keys ...K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
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

func (s *Sorted[K, V]) Range(fn func(key K, value V) bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, k := range s.keys {
		if !fn(k, s.db[k]) {
			return
		}
	}
}

func (s *Sorted[K, V]) Keys() []K {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.keys
}

func (s *Sorted[K, V]) Value(clone bool) map[K]V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if clone {
		return maps.Clone(s.db)
	}
	return s.db
}

func (s *Sorted[K, V]) KVs(clone bool) ([]K, map[K]V) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if clone {
		return slices.Clone(s.keys), maps.Clone(s.db)
	}
	return s.keys, s.db
}

func (s *Sorted[K, V]) Len() int {
	s.mux.RLock()
	val := len(s.keys)
	s.mux.RUnlock()
	return val
}

func (s *Sorted[K, V]) Clear() {
	s.mux.Lock()
	s.keys = nil
	clear(s.db)
	s.mux.Unlock()
}
