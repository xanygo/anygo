//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/xerror"
)

// MemoryCache  是否本地内存缓存
type MemoryCache interface {
	IsMemory() bool
}

// IsMemory 判断一个对象是否 全内存缓存
func IsMemory(c any) bool {
	if c == nil {
		return false
	}
	if nl, ok := c.(MemoryCache); ok {
		return nl.IsMemory()
	}
	return false
}

var _ Cache[string, string] = (*LRU[string, string])(nil)
var _ MCache[string, string] = (*LRU[string, string])(nil)
var _ HasStats = (*LRU[string, string])(nil)
var _ MemoryCache = (*LRU[string, string])(nil)

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

func (lru *LRU[K, V]) IsMemory() bool {
	return true
}

func (lru *LRU[K, V]) Capacity() int {
	return lru.capacity
}

func (lru *LRU[K, V]) Has(ctx context.Context, key K) (bool, error) {
	_, err := lru.Get(ctx, key)
	if err != nil {
		if errors.Is(err, xerror.NotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
	val := el.Value.(*MemValue[K, V])

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
	now := time.Now()
	cacheVal := &MemValue[K, V]{
		Key:      key,
		Data:     value,
		CreateAt: now,
		ExpireAt: now.Add(ttl),
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
	v := el.Value.(*MemValue[K, V])
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
func (lru *LRU[K, V]) Clear() {
	lru.mux.Lock()
	clear(lru.data)
	lru.data = make(map[K]*list.Element, lru.capacity)
	lru.list = list.New()
	lru.mux.Unlock()
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

func (lru *LRU[K, V]) Count() int64 {
	lru.mux.Lock()
	c := len(lru.data)
	lru.mux.Unlock()
	return int64(c)
}

func (lru *LRU[K, V]) Stats() Stats {
	return Stats{
		Read:   lru.readCnt.Load(),
		Write:  lru.writeCnt.Load(),
		Delete: lru.deleteCnt.Load(),
		Hit:    lru.hitCnt.Load(),
	}
}

func (lru *LRU[K, V]) RangeLocked(fn func(item *MemValue[K, V]) (remove bool, goon bool)) {
	lru.mux.Lock()
	defer lru.mux.Unlock()

	for e := lru.list.Back(); e != nil; {
		next := e.Prev()

		kv := e.Value.(*MemValue[K, V])
		remove, goon := fn(kv)
		if remove {
			delete(lru.data, kv.Key)
			lru.list.Remove(e)
			lru.deleteCnt.Add(1)
		}
		if !goon {
			return
		}
		e = next
	}
}

type MemValue[K comparable, V any] struct {
	Key      K
	Data     V
	CreateAt time.Time
	ExpireAt time.Time
}

// Expired 是否已过期
func (v *MemValue[K, V]) Expired() bool {
	return time.Now().After(v.ExpireAt)
}

// NewMemoryFIFO 创建容量满后，淘汰策略为先进先出（FIFO）的内存缓存
func NewMemoryFIFO[K comparable, V any](capacity int) *MemoryXIFO[K, V] {
	return newMemoryFIXO[K, V](capacity, true)
}

// NewMemoryLIFO 创建容量满后，淘汰策略为后进先出（LIFO）的内存缓存
func NewMemoryLIFO[K comparable, V any](capacity int) *MemoryXIFO[K, V] {
	return newMemoryFIXO[K, V](capacity, false)
}

func newMemoryFIXO[K comparable, V any](capacity int, fifo bool) *MemoryXIFO[K, V] {
	if capacity <= 0 {
		panic(fmt.Sprintf("MemoryFIXO with invalid capacity %d", capacity))
	}
	return &MemoryXIFO[K, V]{
		capacity: capacity,
		data:     make(map[K]*list.Element, capacity),
		list:     list.New(),
		mux:      &sync.Mutex{},
		fifo:     fifo,
	}
}

var _ Cache[string, string] = (*MemoryXIFO[string, string])(nil)
var _ MCache[string, string] = (*MemoryXIFO[string, string])(nil)
var _ HasStats = (*MemoryXIFO[string, string])(nil)
var _ MemoryCache = (*MemoryXIFO[string, string])(nil)

// MemoryXIFO 容量满之后，过期策略为 FIFO 或者 LIFO 的 内存缓存
type MemoryXIFO[K comparable, V any] struct {
	capacity int // 容量
	data     map[K]*list.Element
	list     *list.List
	mux      *sync.Mutex
	fifo     bool // 当为 true 时，是 FIFO，为 false 时是 LIFO/FILO

	readCnt   atomic.Uint64
	writeCnt  atomic.Uint64
	deleteCnt atomic.Uint64
	hitCnt    atomic.Uint64
}

func (m *MemoryXIFO[K, V]) IsMemory() bool {
	return true
}

func (m *MemoryXIFO[K, V]) Capacity() int {
	return m.capacity
}

func (m *MemoryXIFO[K, V]) Has(ctx context.Context, key K) (bool, error) {
	_, err := m.Get(ctx, key)
	if err != nil {
		if errors.Is(err, xerror.NotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MemoryXIFO[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	m.readCnt.Add(1)

	m.mux.Lock()
	defer m.mux.Unlock()
	return m.getLocked(key)
}

func (m *MemoryXIFO[K, V]) getLocked(key K) (v V, rr error) {
	el, has := m.data[key]
	if !has {
		return v, xerror.NotFound
	}
	val := el.Value.(*MemValue[K, V])
	if val.Expired() {
		m.list.Remove(el)
		delete(m.data, key)
		return v, xerror.NotFound
	}
	m.hitCnt.Add(1)
	return val.Data, nil
}

func (m *MemoryXIFO[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	m.writeCnt.Add(1)
	m.doSet(key, value, ttl)
	return nil
}

func (m *MemoryXIFO[K, V]) doSet(key K, value V, ttl time.Duration) {
	now := time.Now()
	cacheVal := &MemValue[K, V]{
		Key:      key,
		Data:     value,
		CreateAt: now,
		ExpireAt: now.Add(ttl),
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	el, has := m.data[key]
	if has {
		el.Value = cacheVal
		return
	}
	elm := m.list.PushBack(cacheVal)
	m.data[key] = elm
	if m.list.Len() > m.capacity {
		m.weedOut()
	}
}

func (m *MemoryXIFO[K, V]) weedOut() {
	var el *list.Element
	if m.fifo {
		el = m.list.Front()
	} else {
		el = m.list.Back()
	}
	if el == nil {
		return
	}
	v := el.Value.(*MemValue[K, V])
	delete(m.data, v.Key)
	m.list.Remove(el)
}

func (m *MemoryXIFO[K, V]) Delete(ctx context.Context, keys ...K) error {
	if len(keys) == 0 {
		return nil
	}

	m.deleteCnt.Add(uint64(len(keys)))

	m.mux.Lock()
	defer m.mux.Unlock()
	for _, key := range keys {
		el, has := m.data[key]
		if !has {
			continue
		}
		delete(m.data, key)
		m.list.Remove(el)
	}
	return nil
}

func (m *MemoryXIFO[K, V]) MSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	m.writeCnt.Add(uint64(len(values)))
	for k, v := range values {
		m.doSet(k, v, ttl)
	}
	return nil
}

func (m *MemoryXIFO[K, V]) MGet(ctx context.Context, keys ...K) (map[K]V, error) {
	m.readCnt.Add(uint64(len(keys)))

	m.mux.Lock()
	defer m.mux.Unlock()

	result := make(map[K]V, len(keys))
	for _, key := range keys {
		val, err := m.getLocked(key)
		if err != nil {
			result[key] = val
		}
	}
	return result, nil
}

// Clear 重置、清空所有缓存
func (m *MemoryXIFO[K, V]) Clear() {
	m.mux.Lock()
	clear(m.data)
	m.data = make(map[K]*list.Element, m.capacity)
	m.list = list.New()
	m.mux.Unlock()
}

func (m *MemoryXIFO[K, V]) Keys() []K {
	m.mux.Lock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	m.mux.Unlock()
	return keys
}

func (m *MemoryXIFO[K, V]) Count() int64 {
	m.mux.Lock()
	c := len(m.data)
	m.mux.Unlock()
	return int64(c)
}

func (m *MemoryXIFO[K, V]) Stats() Stats {
	return Stats{
		Keys:   m.Count(),
		Read:   m.readCnt.Load(),
		Write:  m.writeCnt.Load(),
		Delete: m.deleteCnt.Load(),
		Hit:    m.hitCnt.Load(),
	}
}

func (m *MemoryXIFO[K, V]) RangeLocked(fn func(item *MemValue[K, V]) (remove bool, goon bool)) {
	m.mux.Lock()
	defer m.mux.Unlock()

	var e, next *list.Element
	if m.fifo {
		e = m.list.Front()
	} else {
		e = m.list.Back()
	}

	for e != nil {
		if m.fifo {
			next = e.Next()
		} else {
			next = e.Prev()
		}
		kv := e.Value.(*MemValue[K, V])
		remove, goon := fn(kv)
		if remove {
			delete(m.data, kv.Key)
			m.list.Remove(e)
			m.deleteCnt.Add(1)
		}
		if !goon {
			return
		}
		e = next
	}
}
