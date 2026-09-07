//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

import (
	"iter"
	"sync"
)

// Sync 并发安全的 Map,基于 sync.Map 简单封装以支持泛型
type Sync[K comparable, V any] struct {
	storage sync.Map
}

func (m *Sync[K, V]) Load(key K) (value V, ok bool) {
	v1, ok1 := m.storage.Load(key)
	if ok1 {
		v2, ok2 := v1.(V)
		return v2, ok2
	}
	return value, false
}

func (m *Sync[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, ok := m.storage.LoadAndDelete(key)
	if ok {
		return v.(V), true
	}
	return value, false
}

func (m *Sync[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, ok := m.storage.LoadOrStore(key, value)
	return v.(V), ok
}

func (m *Sync[K, V]) Store(key K, value V) {
	m.storage.Store(key, value)
}

func (m *Sync[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	p, ok := m.storage.Swap(key, value)
	if ok {
		return p.(V), true
	}
	return previous, false
}

func (m *Sync[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.storage.CompareAndDelete(key, old)
}

func (m *Sync[K, V]) CompareAndSwap(key K, old V, new V) bool {
	return m.storage.CompareAndSwap(key, old, new)
}

func (m *Sync[K, V]) Delete(key K) {
	m.storage.Delete(key)
}

func (m *Sync[K, V]) DeleteFunc(fn func(key K, value V) bool) {
	m.storage.Range(func(key, value any) bool {
		if fn(key.(K), value.(V)) {
			m.storage.Delete(key)
		}
		return true
	})
}

func (m *Sync[K, V]) Range(fn func(key K, value V) bool) {
	m.storage.Range(func(key, value any) bool {
		return fn(key.(K), value.(V))
	})
}

func (m *Sync[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.storage.Range(func(key, value any) bool {
			return yield(key.(K), value.(V))
		})
	}
}

func (m *Sync[K, V]) Len() int {
	var c int
	m.storage.Range(func(_, _ any) bool {
		c++
		return true
	})
	return c
}

func (m *Sync[K, V]) ToMap() map[K]V {
	val := make(map[K]V)
	m.storage.Range(func(key, value any) bool {
		val[key.(K)] = value.(V)
		return true
	})
	return val
}

func (m *Sync[K, V]) Exists(key K) bool {
	_, ok := m.storage.Load(key)
	return ok
}

func (m *Sync[K, V]) Clear() {
	m.storage.Clear()
}
