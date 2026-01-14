//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-14

package xslice

import "sync"

// Queue 利用 Ring Slice 的队列，非并发安全的
type Queue[T any] struct {
	Capacity int // 最大容量，若 >0 则限制为具体的容量，否则容量无限制

	items []T
	head  int
	tail  int
	size  int
}

// Cap 返回当前 slice 容量
func (r *Queue[T]) Cap() int {
	if r.Capacity > 0 {
		return r.Capacity
	}
	return cap(r.items)
}

// Len 返回队列长度
func (r *Queue[T]) Len() int {
	return r.size
}

// Push 入队
//
//   - 当 Capacity > 0 时，若队列满了则 Push 失败，返回 false
//   - 当 Capacity = 0 时，总是成功，并返回 true
func (r *Queue[T]) Push(v T) bool {
	if r.items == nil {
		if r.Capacity > 0 {
			r.items = make([]T, r.Capacity)
		} else {
			r.items = make([]T, 16)
		}
	}

	if r.size == len(r.items) {
		if r.Capacity > 0 {
			return false
		}
		r.grow()
	}
	r.items[r.tail] = v
	r.tail = (r.tail + 1) % len(r.items)
	r.size++
	return true
}

// Pop 出队
func (r *Queue[T]) Pop() (v T, ok bool) {
	if r.size == 0 {
		return v, false
	}
	v = r.items[r.head]
	var zero T
	r.items[r.head] = zero

	r.head = (r.head + 1) % len(r.items)
	r.size--
	return v, true
}

// First 获取队首的一个元素
func (r *Queue[T]) First() (v T, ok bool) {
	if r.size == 0 {
		return v, false
	}
	return r.items[r.head], true
}

// Discard 丢弃队首的 n 个元素，返回实际丢弃的数量
func (r *Queue[T]) Discard(n int) (discarded int) {
	if r.size == 0 || n < 1 {
		return 0
	}
	var zero T
	total := min(n, r.size)
	for i := 0; i < total; i++ {
		r.items[r.head] = zero
		r.head = (r.head + 1) % len(r.items)
		r.size--
	}
	return total
}

// grow 扩容
func (r *Queue[T]) grow() {
	newCap := len(r.items) * 3 / 2
	if newCap == 0 {
		newCap = 16
	}
	newItems := make([]T, newCap)

	if r.head < r.tail {
		// 数据连续
		copy(newItems, r.items[r.head:r.tail])
	} else {
		// 数据 wrap-around，分两段 copy
		n := copy(newItems, r.items[r.head:])
		copy(newItems[n:], r.items[:r.tail])
	}

	r.head = 0
	r.tail = r.size
	r.items = newItems
}

// SyncQueue 利用 Ring Slice 的队列，并发安全的
type SyncQueue[T any] struct {
	Capacity int // 最大容量，若 >0 则限制为具体的容量，否则容量无限制

	items []T
	head  int
	tail  int
	size  int
	mux   sync.RWMutex
}

// Cap 返回当前 slice 容量
func (r *SyncQueue[T]) Cap() int {
	r.mux.RLock()
	defer r.mux.RUnlock()

	if r.Capacity == 0 {
		return cap(r.items)
	}
	return r.Capacity
}

// Len 返回队列长度
func (r *SyncQueue[T]) Len() int {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return r.size
}

// Push 入队
//
//   - 当 Capacity > 0 时，若队列满了则 Push 失败，返回 false
//   - 当 Capacity = 0 时，总是成功，并返回 true
func (r *SyncQueue[T]) Push(v T) bool {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.items == nil {
		if r.Capacity > 0 {
			r.items = make([]T, r.Capacity)
		} else {
			r.items = make([]T, 16)
		}
	}

	if r.size == len(r.items) {
		if r.Capacity > 0 {
			return false
		}
		r.grow()
	}

	r.items[r.tail] = v
	r.tail = (r.tail + 1) % len(r.items)
	r.size++
	return true
}

// Pop 出队
func (r *SyncQueue[T]) Pop() (v T, ok bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.size == 0 {
		return v, false
	}
	v = r.items[r.head]
	// 可选：清空引用，便于 GC
	var zeroValue T
	r.items[r.head] = zeroValue

	r.head = (r.head + 1) % len(r.items)
	r.size--
	return v, true
}

func (r *SyncQueue[T]) First() (v T, ok bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.size == 0 {
		return v, false
	}
	return r.items[r.head], true
}

// Discard 丢弃队首的 n 个元素，返回实际丢弃的数量
func (r *SyncQueue[T]) Discard(n int) (discarded int) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.size == 0 || n < 1 {
		return 0
	}
	var zero T
	total := min(n, r.size)
	for i := 0; i < total; i++ {
		r.items[r.head] = zero
		r.head = (r.head + 1) % len(r.items)
		r.size--
	}
	return total
}

// grow 扩容
func (r *SyncQueue[T]) grow() {
	newCap := len(r.items) * 3 / 2
	if newCap == 0 {
		newCap = 16
	}
	newItems := make([]T, newCap)

	if r.head < r.tail {
		// 数据连续
		copy(newItems, r.items[r.head:r.tail])
	} else {
		// 数据 wrap-around，分两段 copy
		n := copy(newItems, r.items[r.head:])
		copy(newItems[n:], r.items[:r.tail])
	}

	r.head = 0
	r.tail = r.size
	r.items = newItems
}
