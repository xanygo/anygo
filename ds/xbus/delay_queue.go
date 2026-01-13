//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-09

package xbus

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/xerror"
)

// DelayQueue 全内存的延时队列
type DelayQueue[V any] struct {
	Delay     time.Duration // 延迟时长，应该是 >=0 的时长
	Capacity  int           // 队列最大长度，可选，默认为 0-不限制
	OutBuffer int           // 出栈队列 buffer 长度，可选

	cnt       atomic.Int64
	mu        sync.Mutex
	items     delayItemHeap[V]
	wakeup    chan struct{}
	out       chan V
	startOnce sync.Once
	bgRunning atomic.Bool
	closeOnce sync.Once
	closed    chan struct{}
}

func (q *DelayQueue[V]) start() {
	q.startOnce.Do(func() {
		q.wakeup = make(chan struct{}, 1)
		q.out = make(chan V, q.OutBuffer)
		q.closed = make(chan struct{})

		heap.Init(&q.items)
	})
	select {
	case <-q.closed:
		return
	default:
	}

	if q.bgRunning.CompareAndSwap(false, true) {
		go q.run()
	}
}

func (q *DelayQueue[V]) run() {
	exitTimer := time.NewTimer(10 * time.Second) // 空值此后台携程的最大运行时长
	var waitTimer *time.Timer
	defer func() {
		q.bgRunning.Store(false)
		exitTimer.Stop()
		if waitTimer != nil {
			waitTimer.Stop()
		}
	}()

	for {
		select {
		case <-exitTimer.C:
			return
		case <-q.closed:
			return
		default:
		}

		q.mu.Lock()

		if q.items.Len() == 0 {
			q.mu.Unlock()
			select {
			case <-q.wakeup:
				continue
			case <-exitTimer.C:
				return
			case <-q.closed:
				return
			}
		}

		it := q.items[0]
		now := time.Now()

		if it.readyAt.After(now) { // item 还没有到底延迟的时间
			d := it.readyAt.Sub(now) // 需要等待此时长，才可以被 POP
			q.mu.Unlock()

			if waitTimer == nil {
				waitTimer = time.NewTimer(d)
			} else {
				waitTimer.Reset(d)
			}

			select {
			case <-waitTimer.C:
			case <-q.wakeup:
			}
			continue
		}

		heap.Pop(&q.items)
		q.mu.Unlock()

		select {
		case q.out <- it.value:
		case <-q.closed:
			return
		}
	}
}

// Push 往队列里添加一个元素，
//
// 返回值：
//   - true-添加成功
//   - false-失败,可能是 Queue 已被Stop、容量满
func (q *DelayQueue[V]) Push(value V) bool {
	q.start()

	select {
	case <-q.closed:
		return false
	default:
	}

	item := &delayItem[V]{
		value:   value,
		readyAt: time.Now().Add(q.Delay),
	}
	q.mu.Lock()
	if q.Capacity > 0 && len(q.items) >= q.Capacity {
		q.mu.Unlock()
		return false
	}
	heap.Push(&q.items, item)
	q.cnt.Add(1)
	q.mu.Unlock()

	select {
	case q.wakeup <- struct{}{}:
	default:
	}
	return true
}

// TryPop 同步的，若有数据则返回 v,true， 没有这返回 v,false
func (q *DelayQueue[V]) TryPop() (v V, ok bool) {
	q.start()

	select {
	case v = <-q.out:
		q.cnt.Add(-1)
		return v, true
	default:
		return v, false
	}
}

func (q *DelayQueue[V]) PopWait() (v V, err error) {
	return q.Pop(context.Background())
}

func (q *DelayQueue[V]) Pop(ctx context.Context) (v V, err error) {
	q.start()

	select {
	case v = <-q.out:
		q.cnt.Add(-1)
		return v, nil
	case <-ctx.Done():
		return v, ctx.Err()
	case <-q.closed:
		return v, xerror.Closed
	}
}

// Len 返回队列里总共的元素个数
func (q *DelayQueue[V]) Len() int {
	return int(q.cnt.Load())
}

func (q *DelayQueue[V]) Stop() {
	q.start()
	q.closeOnce.Do(func() {
		close(q.closed)
	})
}

// DeleteByFunc 删除，会整体加锁。若 OutBuffer > 0,在出栈队列里的不会删除
func (q *DelayQueue[V]) DeleteByFunc(delFn func(v V) bool) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 先统计删除元素数量
	var count int
	for _, it := range q.items {
		if delFn(it.value) {
			count++
		}
	}

	if count == 0 {
		return 0
	}

	if count*2 >= len(q.items) {
		newItems := make(delayItemHeap[V], 0, len(q.items)-count)
		for _, it := range q.items {
			if delFn(it.value) {
				q.cnt.Add(-1)
			} else {
				newItems = append(newItems, it)
			}
		}
		q.items = newItems
		heap.Init(&q.items)
		return count
	}

	var i int
	for i < len(q.items) {
		if delFn(q.items[i].value) {
			q.cnt.Add(-1)
			heap.Remove(&q.items, i)
			// heap.Remove 会将最后一个元素放到 i 位置，所以 i 不变
			continue
		}
		i++
	}

	return count
}

type delayItem[V any] struct {
	value   V
	readyAt time.Time
	index   int
}

type delayItemHeap[V any] []*delayItem[V]

func (h delayItemHeap[V]) Len() int { return len(h) }

func (h delayItemHeap[V]) Less(i, j int) bool {
	return h[i].readyAt.Before(h[j].readyAt)
}

func (h delayItemHeap[V]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *delayItemHeap[V]) Push(x any) {
	it := x.(*delayItem[V])
	it.index = len(*h)
	*h = append(*h, it)
}

func (h *delayItemHeap[V]) Pop() any {
	old := *h
	n := len(old)
	it := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return it
}
