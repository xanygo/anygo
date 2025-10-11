//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package xcounter

import (
	"sync"
	"time"
)

// NewSliding 创建一个滑动窗口计数器
func NewSliding(window, resolution time.Duration) *Sliding {
	if resolution <= 0 {
		resolution = time.Second
	}
	if window < resolution {
		window = resolution
	}
	size := int(window / resolution)
	return &Sliding{
		resolution: resolution,
		window:     window,
		buckets:    make([]bucket1, size),
		start:      time.Now(),
	}
}

// Sliding 滑动窗口计数器
type Sliding struct {
	mu         sync.Mutex
	resolution time.Duration // 每个桶的时间粒度
	window     time.Duration // 总窗口时长
	buckets    []bucket1
	start      time.Time
}

type bucket1 struct {
	ts    time.Time
	count int64
}

func (c *Sliding) Incr() {
	c.IncrN(1)
}

// IncrN 往计数器加 N
func (c *Sliding) IncrN(n int64) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	index := c.indexFor(now)

	if !c.sameSlot(c.buckets[index].ts, now) {
		// 重置过期的桶
		c.buckets[index].ts = now.Truncate(c.resolution)
		c.buckets[index].count = 0
	}
	c.buckets[index].count += n
}

// Total 返回整个窗口的计数
func (c *Sliding) Total() int64 {
	return c.Count(c.window)
}

// Count 返回最近 duration 内的计数
func (c *Sliding) Count(d time.Duration) int64 {
	now := time.Now()
	if d > c.window {
		d = c.window
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	var sum int64
	for i := 0; i < len(c.buckets); i++ {
		b := c.buckets[i]
		if now.Sub(b.ts) < d {
			sum += b.count
		}
	}
	return sum
}

func (c *Sliding) CountWindowN(n int) int64 {
	return c.Count(c.window * time.Duration(n))
}

func (c *Sliding) CountWindow() int64 {
	return c.CountWindowN(1)
}

// indexFor 计算当前时间对应的桶索引
func (c *Sliding) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.resolution) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *Sliding) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.resolution).Equal(t2.Truncate(c.resolution))
}
