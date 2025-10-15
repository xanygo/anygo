//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/xerror"
)

var _ Cache[string, string] = (*LRU[string, string])(nil)
var _ MCache[string, string] = (*LRU[string, string])(nil)
var _ HasStats = (*LRU[string, string])(nil)

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

// LRU 最近最少使用( Least Recently Used ) 全内存缓存组件。
type LRU[K comparable, V any] struct {
	capacity int // 容量
	data     map[K]*list.Element
	list     *list.List
	mux      *sync.Mutex

	readCnt   atomic.Uint64
	writeCnt  atomic.Uint64
	deleteCnt atomic.Uint64
	hitCnt    atomic.Uint64
}

func (lru *LRU[K, V]) Get(_ context.Context, key K) (v V, err error) {
	lru.readCnt.Add(1)

	lru.mux.Lock()
	defer lru.mux.Unlock()
	return lru.getLocked(key)
}

func (lru *LRU[K, V]) getLocked(key K) (v V, rr error) {
	el, has := lru.data[key]
	if !has {
		return v, xerror.NotFound
	}
	val := el.Value.(*lruValue[K, V])

	if val.Expired() {
		lru.list.Remove(el)
		delete(lru.data, key)
		return v, xerror.NotFound
	}
	lru.hitCnt.Add(1)
	lru.list.MoveToFront(el)
	return val.Data, nil
}

func (lru *LRU[K, V]) MGet(_ context.Context, keys ...K) (map[K]V, error) {
	lru.readCnt.Add(uint64(len(keys)))

	lru.mux.Lock()
	defer lru.mux.Unlock()

	result := make(map[K]V, len(keys))
	for _, key := range keys {
		val, err := lru.getLocked(key)
		if err != nil {
			result[key] = val
		}
	}
	return result, nil
}

func (lru *LRU[K, V]) Set(_ context.Context, key K, value V, ttl time.Duration) error {
	lru.writeCnt.Add(1)
	lru.doSet(key, value, ttl)
	return nil
}

func (lru *LRU[K, V]) doSet(key K, value V, ttl time.Duration) {
	cacheVal := &lruValue[K, V]{
		Key:      key,
		Data:     value,
		ExpireAt: time.Now().Add(ttl),
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

func (lru *LRU[K, V]) MSet(_ context.Context, values map[K]V, ttl time.Duration) error {
	lru.writeCnt.Add(uint64(len(values)))
	for k, v := range values {
		lru.doSet(k, v, ttl)
	}
	return nil
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

func (lru *LRU[K, V]) Delete(ctx context.Context, keys ...K) error {
	lru.deleteCnt.Add(uint64(len(keys)))

	if len(keys) == 0 {
		return nil
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
	return nil
}

// Clear 重置、清空所有缓存
func (lru *LRU[K, V]) Clear(ctx context.Context) error {
	lru.mux.Lock()
	clear(lru.data)
	lru.data = make(map[K]*list.Element, lru.capacity)
	lru.list = list.New()
	lru.mux.Unlock()
	return nil
}

func (lru *LRU[K, V]) Keys() []K {
	lru.mux.Lock()
	keys := make([]K, 0, len(lru.data))
	for k := range lru.data {
		keys = append(keys, k)
	}
	lru.mux.Unlock()
	return keys
}

func (lru *LRU[K, V]) Count() int {
	lru.mux.Lock()
	c := len(lru.data)
	lru.mux.Unlock()
	return c
}

func (lru *LRU[K, V]) Stats() Stats {
	return Stats{
		Read:   lru.readCnt.Load(),
		Write:  lru.writeCnt.Load(),
		Delete: lru.deleteCnt.Load(),
		Hit:    lru.hitCnt.Load(),
	}
}

type lruValue[K comparable, V any] struct {
	Key      K
	Data     V
	ExpireAt time.Time
}

// Expired 是否已过期
func (v *lruValue[K, V]) Expired() bool {
	return time.Now().After(v.ExpireAt)
}
