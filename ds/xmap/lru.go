//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-10

package xmap

import (
	"container/list"
	"fmt"
	"sync"
)

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity <= 0 {
		panic(fmt.Sprintf("NewLRU with invalid capacity %d", capacity))
	}
	return &LRU[K, V]{
		capacity: capacity,
		data:     make(map[K]*list.Element, capacity),
		list:     list.New(),
		mux:      &sync.Mutex{},
	}
}

// LRU 最近最少使用( Least Recently Used ) 全内存缓存组件
type LRU[K comparable, V any] struct {
	capacity int // 容量
	data     map[K]*list.Element
	list     *list.List
	mux      *sync.Mutex
}

func (lru *LRU[K, V]) Get(key K) (v V, ok bool) {
	lru.mux.Lock()
	defer lru.mux.Unlock()
	el, has := lru.data[key]
	if !has {
		return v, false
	}
	val := el.Value.(*lruValue[K, V])
	lru.list.MoveToFront(el)
	return val.Data, true
}

func (lru *LRU[K, V]) Set(key K, value V) {
	cacheVal := &lruValue[K, V]{
		Key:  key,
		Data: value,
	}
	lru.mux.Lock()
	defer lru.mux.Unlock()
	el, has := lru.data[key]
	if has {
		el.Value = cacheVal
		lru.list.MoveToFront(el)
		return
	}
	elm := lru.list.PushFront(cacheVal)
	lru.data[key] = elm
	if lru.list.Len() > lru.capacity {
		lru.weedOut()
	}
}

func (lru *LRU[K, V]) weedOut() {
	el := lru.list.Back()
	if el == nil {
		return
	}
	v := el.Value.(*lruValue[K, V])
	delete(lru.data, v.Key)
	lru.list.Remove(el)
}

func (lru *LRU[K, V]) Delete(keys ...K) {
	if len(keys) == 0 {
		return
	}
	lru.mux.Lock()
	defer lru.mux.Unlock()
	for _, key := range keys {
		el, has := lru.data[key]
		if !has {
			continue
		}
		delete(lru.data, key)
		lru.list.Remove(el)
	}
}

// Clear 重置、清空所有缓存
func (lru *LRU[K, V]) Clear() {
	lru.mux.Lock()
	clear(lru.data)
	lru.data = make(map[K]*list.Element, lru.capacity)
	lru.list = list.New()
	lru.mux.Unlock()
}

// Keys 返回所有的 key，已按照使用时间排序，最近使用的排在前面
func (lru *LRU[K, V]) Keys() []K {
	lru.mux.Lock()
	keys := make([]K, 0, len(lru.data))
	for el := lru.list.Front(); el != nil; el = el.Next() {
		val := el.Value.(*lruValue[K, V])
		keys = append(keys, val.Key)
	}
	lru.mux.Unlock()
	return keys
}

func (lru *LRU[K, V]) Map() map[K]V {
	lru.mux.Lock()
	result := make(map[K]V, len(lru.data))
	for _, v := range lru.data {
		item := v.Value.(*lruValue[K, V])
		result[item.Key] = item.Data
	}
	lru.mux.Unlock()
	return result
}

func (lru *LRU[K, V]) Len() int {
	lru.mux.Lock()
	c := len(lru.data)
	lru.mux.Unlock()
	return c
}

type lruValue[K comparable, V any] struct {
	Key  K
	Data V
}

type LRUReader[K string, V any] struct {
	New   func(key K) V
	Store *LRU[K, V]
	mux   sync.Mutex
}

func (lr *LRUReader[K, V]) Get(key K) V {
	if val, ok := lr.Store.Get(key); ok {
		return val
	}
	lr.mux.Lock()
	defer lr.mux.Unlock()
	val := lr.New(key)
	lr.Store.Set(key, val)
	return val
}

func (lr *LRUReader[K, V]) Delete(keys ...K) {
	lr.Store.Delete(keys...)
}
