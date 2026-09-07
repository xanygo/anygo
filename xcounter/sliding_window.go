//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package xcounter

import (
	"sync"
	"time"
)

// NewSlidingWindow 创建一个滑动窗口计数器
// windowSize: 总时间长度
// bucketSize: 单个区间大小
func NewSlidingWindow(windowSize, bucketDuration time.Duration) *SlidingWindow {
	if bucketDuration <= 0 {
		bucketDuration = time.Second
	}
	if windowSize < bucketDuration {
		windowSize = bucketDuration
	}
	size := int(windowSize / bucketDuration)
	return &SlidingWindow{
		bucketSize: bucketDuration,
		windowSize: windowSize,
		buckets:    make([]bucket1, size),
		start:      time.Now(),
	}
}

// SlidingWindow 滑动窗口计数器
type SlidingWindow struct {
	mu         sync.Mutex
	bucketSize time.Duration // 每个桶的时间粒度
	windowSize time.Duration // 总窗口时长
	buckets    []bucket1
	start      time.Time
	total      int64 // 累计计数，不会随着窗口滑动清零、重置
}

type bucket1 struct {
	ts    time.Time
	count int64
}

func (c *SlidingWindow) Start() time.Time {
	return c.start
}

func (c *SlidingWindow) WindowSize() time.Duration {
	return c.windowSize
}

func (c *SlidingWindow) BucketSize() time.Duration {
	return c.bucketSize
}

func (c *SlidingWindow) Incr() {
	c.IncrN(1)
}

// IncrN 往计数器加 N
func (c *SlidingWindow) IncrN(n int64) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.total += n

	index := c.indexFor(now)

	if !c.sameSlot(c.buckets[index].ts, now) {
		// 重置过期的桶
		c.buckets[index].ts = now.Truncate(c.bucketSize)
		c.buckets[index].count = 0
	}
	c.buckets[index].count += n
}

// LifetimeTotal 返回累计计数
func (c *SlidingWindow) LifetimeTotal() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

// WindowTotal 返回整个窗口的计数
func (c *SlidingWindow) WindowTotal() int64 {
	return c.Count(c.windowSize)
}

// Count 返回最近 duration 内的计数
func (c *SlidingWindow) Count(d time.Duration) int64 {
	now := time.Now()
	if d > c.windowSize {
		d = c.windowSize
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

func (c *SlidingWindow) CountWindowN(n int) int64 {
	return c.Count(c.windowSize * time.Duration(n))
}

func (c *SlidingWindow) CountWindow() int64 {
	return c.CountWindowN(1)
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingWindow) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.bucketSize) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingWindow) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.bucketSize).Equal(t2.Truncate(c.bucketSize))
}
