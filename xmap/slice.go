//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xmap

import (
	"maps"
	"sync"

	"github.com/xanygo/anygo/internal/zslice"
)

// Slice 值为 slice 的 并发安全的 map ( map[K][]V )
type Slice[K, V comparable] struct {
	data map[K][]V
	mux  sync.RWMutex
}

func (s *Slice[K, V]) Set(key K, values ...V) {
	s.mux.Lock()
	if s.data == nil {
		s.data = make(map[K][]V)
	}
	s.data[key] = values
	s.mux.Unlock()
}

func (s *Slice[K, V]) Get(key K) []V {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 {
		return nil
	}
	return s.data[key]
}

func (s *Slice[K, V]) GetFirst(key K) (v V) {
	vs := s.Get(key)
	if len(vs) == 0 {
		return v
	}
	return vs[0]
}

func (s *Slice[K, V]) Add(key K, values ...V) {
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

func (s *Slice[K, V]) Delete(keys ...K) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if len(s.data) == 0 {
		return
	}
	for _, key := range keys {
		delete(s.data, key)
	}
}

func (s *Slice[K, V]) DeleteValue(key K, values ...V) {
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

func (s *Slice[K, V]) HasKey(key K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 {
		return false
	}

	_, has := s.data[key]
	return has
}

func (s *Slice[K, V]) HasValue(key K, values ...V) bool {
	if len(values) == 0 {
		return false
	}
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 {
		return false
	}

	return zslice.ContainsAny(s.data[key], values...)
}

func (s *Slice[K, V]) Keys() []K {
	s.mux.RLock()
	result := Keys(s.data)
	s.mux.RUnlock()
	return result
}

func (s *Slice[K, V]) Map() map[K][]V {
	s.mux.RLock()
	result := maps.Clone(s.data)
	s.mux.RUnlock()
	return result
}
