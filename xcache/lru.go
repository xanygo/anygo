//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xanygo/anygo/xerror"
)

var _ Cache[string, string] = (*LRU[string, string])(nil)

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
}

func (lru *LRU[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	lru.mux.Lock()
	defer lru.mux.Unlock()
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
	lru.list.MoveToFront(el)
	return val.Data, nil
}

func (lru *LRU[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
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
		return nil
	}
	elm := lru.list.PushFront(cacheVal)
	lru.data[key] = elm
	if lru.list.Len() > lru.capacity {
		lru.weedOut()
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

type lruValue[K comparable, V any] struct {
	Key      K
	Data     V
	ExpireAt time.Time
}

// Expired 是否已过期
func (v *lruValue[K, V]) Expired() bool {
	return time.Now().After(v.ExpireAt)
}
