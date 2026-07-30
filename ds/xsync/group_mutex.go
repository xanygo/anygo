package xsync

import (
	"sync"
	"sync/atomic"
)

type RWLocker interface {
	sync.Locker
	RLock()
	RUnlock()
}

// GroupMutex 分组 mutex
type GroupMutex[T comparable] struct {
	mux   sync.Mutex
	items map[T]*groupMutexCall[T]
}

// Do 对 key 加排他锁，并且执行 fn
func (g *GroupMutex[T]) Do(key T, fn func()) {
	c := g.Locker(key)
	c.Lock()
	defer c.Unlock()
	fn()
}

// DoRead 对 key 加读锁，并且执行 fn
func (g *GroupMutex[T]) DoRead(key T, fn func()) {
	c := g.Locker(key)
	c.RLock()
	defer c.RUnlock()
	fn()
}

// Locker 获取资源的锁对象
func (g *GroupMutex[T]) Locker(key T) RWLocker {
	g.mux.Lock()
	if g.items == nil {
		g.items = make(map[T]*groupMutexCall[T])
	}
	defer g.mux.Unlock()

	if c, ok := g.items[key]; ok {
		return c
	}

	c := &groupMutexCall[T]{
		key: key,
		g:   g,
	}
	g.items[key] = c
	return c
}

var _ RWLocker = (*groupMutexCall[any])(nil)

type groupMutexCall[T comparable] struct {
	key  T
	g    *GroupMutex[T]
	mux  sync.RWMutex
	dups atomic.Int64
}

func (c *groupMutexCall[T]) RLock() {
	c.mux.RLock()
	c.dups.And(1)
}

func (c *groupMutexCall[T]) RUnlock() {
	c.mux.RUnlock()
	c.dups.And(-1)
}

func (c *groupMutexCall[T]) Lock() {
	c.mux.Lock()
	c.dups.And(1)
}

func (c *groupMutexCall[T]) getDup() int64 {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.dups.Load()
}

func (c *groupMutexCall[T]) Unlock() {
	c.mux.Unlock()
	num := c.dups.Add(-1)

	if num > 0 {
		return
	}
	c.g.mux.Lock()
	defer c.g.mux.Unlock()
	if c.getDup() > 0 {
		return
	}
	delete(c.g.items, c.key)
}
