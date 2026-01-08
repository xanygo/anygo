//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-07

package xcounter

import (
	"sync"
	"time"
)

// NewSlidingWindowStats 创建一个滑动窗口计数器
func NewSlidingWindowStats(windowSize, bucketDuration time.Duration) *SlidingWindowStats {
	if bucketDuration <= 0 {
		bucketDuration = time.Second
	}
	if windowSize < bucketDuration {
		windowSize = bucketDuration
	}
	size := int(windowSize / bucketDuration)
	return &SlidingWindowStats{
		bucketSize: bucketDuration,
		windowSize: windowSize,
		buckets:    make([]bucket2, size),
		start:      time.Now(),
	}
}

// SlidingWindowStats 可同时记录成功数和失败数的滑动窗口计数器
type SlidingWindowStats struct {
	mu           sync.Mutex
	bucketSize   time.Duration // 每个桶的时间粒度
	windowSize   time.Duration // 总窗口时长
	buckets      []bucket2
	start        time.Time
	totalSuccess int64
	totalFail    int64
}

type bucket2 struct {
	ts      time.Time
	success int64
	failure int64
}

func (c *SlidingWindowStats) Start() time.Time {
	return c.start
}

func (c *SlidingWindowStats) WindowSize() time.Duration {
	return c.windowSize
}

func (c *SlidingWindowStats) BucketSize() time.Duration {
	return c.bucketSize
}

func (c *SlidingWindowStats) IncrAuto(err error) {
	if err == nil {
		c.IncrN(1, 0)
	} else {
		c.IncrN(0, 1)
	}
}

// IncrN 往计数器加 N
func (c *SlidingWindowStats) IncrN(success, failure int64) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalSuccess += success
	c.totalFail += failure

	index := c.indexFor(now)

	if !c.sameSlot(c.buckets[index].ts, now) {
		// 重置过期的桶
		c.buckets[index].ts = now.Truncate(c.bucketSize)
		c.buckets[index].success = 0
		c.buckets[index].failure = 0
	}
	c.buckets[index].success += success
	c.buckets[index].failure += failure
}

// LifetimeCounts 返回从创建开始，累计的计数
func (c *SlidingWindowStats) LifetimeCounts() (success, failure int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSuccess, c.totalFail
}

// LifetimeTotal 返回从创建开始，累计的计数
func (c *SlidingWindowStats) LifetimeTotal() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSuccess + c.totalFail
}

// LifetimeSuccess 返回整个窗口的成功数
func (c *SlidingWindowStats) LifetimeSuccess() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSuccess
}

// LifetimeFailure 返回从创建开始的失败数
func (c *SlidingWindowStats) LifetimeFailure() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalFail
}

// WindowCounts 返回整个窗口内的计数
func (c *SlidingWindowStats) WindowCounts() (success, failure int64) {
	return c.Counts(c.windowSize)
}

// WindowTotal 返回整个窗口的计数: 成功数 + 失败数
func (c *SlidingWindowStats) WindowTotal() int64 {
	success, failure := c.WindowCounts()
	return success + failure
}

// WindowSuccess 返回整个窗口的成功数
func (c *SlidingWindowStats) WindowSuccess() int64 {
	success, _ := c.WindowCounts()
	return success
}

// WindowFailure 返回整个窗口的失败数
func (c *SlidingWindowStats) WindowFailure() int64 {
	_, failure := c.WindowCounts()
	return failure
}

// Counts 返回最近 duration 内的计数
func (c *SlidingWindowStats) Counts(d time.Duration) (success, failure int64) {
	now := time.Now()
	if d > c.windowSize {
		d = c.windowSize
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < len(c.buckets); i++ {
		b := c.buckets[i]
		if now.Sub(b.ts) < d {
			success += b.success
			failure += b.failure
		}
	}
	return success, failure
}

// Count 获取时间段内的 成功数 + 失败数
func (c *SlidingWindowStats) Count(d time.Duration) int64 {
	success, failure := c.Counts(d)
	return success + failure
}

func (c *SlidingWindowStats) Success(d time.Duration) int64 {
	success, _ := c.Counts(d)
	return success
}

func (c *SlidingWindowStats) Failure(d time.Duration) int64 {
	_, failure := c.Counts(d)
	return failure
}

// 最近 N 窗口统计

func (c *SlidingWindowStats) WindowNCounts(n int) (success, failure int64) {
	return c.Counts(c.windowSize * time.Duration(n))
}

func (c *SlidingWindowStats) WindowNTotal(n int) int64 {
	success, failure := c.WindowNCounts(n)
	return success + failure
}

func (c *SlidingWindowStats) WindowNSuccess(n int) int64 {
	success, _ := c.WindowNCounts(n)
	return success
}

func (c *SlidingWindowStats) WindowNFailure(n int) int64 {
	_, failure := c.WindowNCounts(n)
	return failure
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingWindowStats) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.bucketSize) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingWindowStats) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.bucketSize).Equal(t2.Truncate(c.bucketSize))
}
