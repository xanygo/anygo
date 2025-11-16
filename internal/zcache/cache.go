//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package zcache

import "sync"

type MapCache[K any, V any] struct {
	New func(key K) V
	db  sync.Map
}

func (mc *MapCache[K, V]) Load(key K) (value V, ok bool) {
	v, ok := mc.db.Load(key)
	if ok {
		return v.(V), ok
	}
	return value, false
}

func (mc *MapCache[K, V]) Count() int {
	var count int
	mc.db.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

func (mc *MapCache[K, V]) Clear() {
	mc.db.Clear()
}

func (mc *MapCache[K, V]) Set(key K, value V) {
	mc.db.Store(key, value)
}

func (mc *MapCache[K, V]) Get1(key K) V {
	value, ok := mc.db.Load(key)
	if ok {
		return value.(V)
	}
	nv := mc.New(key)
	mc.db.Store(key, nv)
	return nv
}

func (mc *MapCache[K, V]) Get2(key K, fn func(key K) V) V {
	value, ok := mc.db.Load(key)
	if ok {
		return value.(V)
	}
	nv := fn(key)
	mc.db.Store(key, nv)
	return nv
}
