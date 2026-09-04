package xsync

import (
	"sync"
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
//
//	获取后必须使用，在调用 Unlock 和 RUnlock 时会检查引用次数，
//	当引用次数为 0 后，会从 GroupMutex 删除引用
func (g *GroupMutex[T]) Locker(key T) RWLocker {
	g.mux.Lock()
	if g.items == nil {
		g.items = make(map[T]*groupMutexCall[T])
	}
	defer g.mux.Unlock()

	if c, ok := g.items[key]; ok {
		c.dups++
		return c
	}

	c := &groupMutexCall[T]{
		key: key,
		g:   g,
	}
	g.items[key] = c
	c.dups++
	return c
}

var _ RWLocker = (*groupMutexCall[any])(nil)

type groupMutexCall[T comparable] struct {
	key  T
	g    *GroupMutex[T]
	mux  sync.RWMutex
	dups int // 使用 g 的 mux 锁
}

func (c *groupMutexCall[T]) RLock() {
	c.mux.RLock()
}

func (c *groupMutexCall[T]) RUnlock() {
	c.mux.RUnlock()
	c.free()
}

func (c *groupMutexCall[T]) Lock() {
	c.mux.Lock()
}

func (c *groupMutexCall[T]) free() {
	c.g.mux.Lock()
	defer c.g.mux.Unlock()
	c.dups--
	if c.dups == 0 {
		delete(c.g.items, c.key)
	}
}

func (c *groupMutexCall[T]) Unlock() {
	c.mux.Unlock()
	c.free()
}
