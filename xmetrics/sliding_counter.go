//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package xmetrics

import (
	"sync"
	"time"
)

// NewSlidingCounter 创建一个滑动窗口计数器
func NewSlidingCounter(window, resolution time.Duration) *SlidingCounter {
	if resolution <= 0 {
		resolution = time.Second
	}
	if window < resolution {
		window = resolution
	}
	size := int(window / resolution)
	return &SlidingCounter{
		resolution: resolution,
		window:     window,
		buckets:    make([]bucket1, size),
		start:      time.Now(),
	}
}

// SlidingCounter 滑动窗口计数器
type SlidingCounter struct {
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

func (c *SlidingCounter) Incr() {
	c.IncrN(1)
}

// IncrN 往计数器加 N
func (c *SlidingCounter) IncrN(n int64) {
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
func (c *SlidingCounter) Total() int64 {
	return c.Count(c.window)
}

// Count 返回最近 duration 内的计数
func (c *SlidingCounter) Count(d time.Duration) int64 {
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

func (c *SlidingCounter) CountWindowN(n int) int64 {
	return c.Count(c.window * time.Duration(n))
}

func (c *SlidingCounter) CountWindow() int64 {
	return c.CountWindowN(1)
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingCounter) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.resolution) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingCounter) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.resolution).Equal(t2.Truncate(c.resolution))
}

// NewSlidingDualCounter 创建一个滑动窗口计数器
func NewSlidingDualCounter(window, resolution time.Duration) *SlidingDualCounter {
	if resolution <= 0 {
		resolution = time.Second
	}
	if window < resolution {
		window = resolution
	}
	size := int(window / resolution)
	return &SlidingDualCounter{
		resolution: resolution,
		window:     window,
		buckets:    make([]bucket2, size),
		start:      time.Now(),
	}
}

// SlidingDualCounter 可同时记录成功数和失败数的滑动窗口计数器
type SlidingDualCounter struct {
	mu         sync.Mutex
	resolution time.Duration // 每个桶的时间粒度
	window     time.Duration // 总窗口时长
	buckets    []bucket2
	start      time.Time
}

type bucket2 struct {
	ts      time.Time
	success int64
	failure int64
}

func (c *SlidingDualCounter) IncrAuto(err error) {
	if err == nil {
		c.IncrN(1, 0)
	} else {
		c.IncrN(0, 1)
	}
}

// IncrN 往计数器加 N
func (c *SlidingDualCounter) IncrN(success, failure int64) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	index := c.indexFor(now)

	if !c.sameSlot(c.buckets[index].ts, now) {
		// 重置过期的桶
		c.buckets[index].ts = now.Truncate(c.resolution)
		c.buckets[index].success = 0
		c.buckets[index].failure = 0
	}
	c.buckets[index].success += success
	c.buckets[index].failure += failure
}

// TotalDual 返回整个窗口的计数
func (c *SlidingDualCounter) TotalDual() (success, failure int64) {
	return c.CountDual(c.window)
}

// Total 返回整个窗口的计数: 成功数 + 失败数
func (c *SlidingDualCounter) Total() int64 {
	success, failure := c.TotalDual()
	return success + failure
}

// TotalSuccess 返回整个窗口的成功数
func (c *SlidingDualCounter) TotalSuccess() int64 {
	success, _ := c.TotalDual()
	return success
}

// TotalFailure 返回整个窗口的失败数
func (c *SlidingDualCounter) TotalFailure() int64 {
	_, failure := c.TotalDual()
	return failure
}

func (c *SlidingDualCounter) CountSuccess(d time.Duration) int64 {
	success, _ := c.CountDual(d)
	return success
}

func (c *SlidingDualCounter) CountFailure(d time.Duration) int64 {
	_, failure := c.CountDual(d)
	return failure
}

// CountDual 返回最近 duration 内的计数
func (c *SlidingDualCounter) CountDual(d time.Duration) (success, failure int64) {
	now := time.Now()
	if d > c.window {
		d = c.window
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
func (c *SlidingDualCounter) Count(d time.Duration) int64 {
	success, failure := c.CountDual(d)
	return success + failure
}

func (c *SlidingDualCounter) CountDualWindowN(n int) (success, failure int64) {
	return c.CountDual(c.window * time.Duration(n))
}

func (c *SlidingDualCounter) CountSuccessWindowN(n int) int64 {
	success, _ := c.CountDualWindowN(n)
	return success
}

func (c *SlidingDualCounter) CountFailureWindowN(n int) int64 {
	_, failure := c.CountDualWindowN(n)
	return failure
}

func (c *SlidingDualCounter) CountWindowN(n int) int64 {
	success, failure := c.CountDualWindowN(n)
	return success + failure
}

func (c *SlidingDualCounter) CountDualWindow() (success, failure int64) {
	return c.CountDualWindowN(1)
}

func (c *SlidingDualCounter) CountWindow() int64 {
	success, failure := c.CountDualWindow()
	return success + failure
}

func (c *SlidingDualCounter) CountSuccessWindow() int64 {
	success, _ := c.CountDualWindow()
	return success
}

func (c *SlidingDualCounter) CountFailureWindow() int64 {
	_, failure := c.CountDualWindow()
	return failure
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingDualCounter) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.resolution) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingDualCounter) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.resolution).Equal(t2.Truncate(c.resolution))
}
