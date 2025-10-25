//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-07

package xcounter

import (
	"sync"
	"time"
)

// NewSlidingDual 创建一个滑动窗口计数器
func NewSlidingDual(window, interval time.Duration) *SlidingDual {
	if interval <= 0 {
		interval = time.Second
	}
	if window < interval {
		window = interval
	}
	size := int(window / interval)
	return &SlidingDual{
		resolution: interval,
		window:     window,
		buckets:    make([]bucket2, size),
		start:      time.Now(),
	}
}

// SlidingDual 可同时记录成功数和失败数的滑动窗口计数器
type SlidingDual struct {
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

func (c *SlidingDual) IncrAuto(err error) {
	if err == nil {
		c.IncrN(1, 0)
	} else {
		c.IncrN(0, 1)
	}
}

// IncrN 往计数器加 N
func (c *SlidingDual) IncrN(success, failure int64) {
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
func (c *SlidingDual) TotalDual() (success, failure int64) {
	return c.CountDual(c.window)
}

// Total 返回整个窗口的计数: 成功数 + 失败数
func (c *SlidingDual) Total() int64 {
	success, failure := c.TotalDual()
	return success + failure
}

// TotalSuccess 返回整个窗口的成功数
func (c *SlidingDual) TotalSuccess() int64 {
	success, _ := c.TotalDual()
	return success
}

// TotalFailure 返回整个窗口的失败数
func (c *SlidingDual) TotalFailure() int64 {
	_, failure := c.TotalDual()
	return failure
}

func (c *SlidingDual) CountSuccess(d time.Duration) int64 {
	success, _ := c.CountDual(d)
	return success
}

func (c *SlidingDual) CountFailure(d time.Duration) int64 {
	_, failure := c.CountDual(d)
	return failure
}

// CountDual 返回最近 duration 内的计数
func (c *SlidingDual) CountDual(d time.Duration) (success, failure int64) {
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
func (c *SlidingDual) Count(d time.Duration) int64 {
	success, failure := c.CountDual(d)
	return success + failure
}

func (c *SlidingDual) CountDualWindowN(n int) (success, failure int64) {
	return c.CountDual(c.window * time.Duration(n))
}

func (c *SlidingDual) CountSuccessWindowN(n int) int64 {
	success, _ := c.CountDualWindowN(n)
	return success
}

func (c *SlidingDual) CountFailureWindowN(n int) int64 {
	_, failure := c.CountDualWindowN(n)
	return failure
}

func (c *SlidingDual) CountWindowN(n int) int64 {
	success, failure := c.CountDualWindowN(n)
	return success + failure
}

func (c *SlidingDual) CountDualWindow() (success, failure int64) {
	return c.CountDualWindowN(1)
}

func (c *SlidingDual) CountWindow() int64 {
	success, failure := c.CountDualWindow()
	return success + failure
}

func (c *SlidingDual) CountSuccessWindow() int64 {
	success, _ := c.CountDualWindow()
	return success
}

func (c *SlidingDual) CountFailureWindow() int64 {
	_, failure := c.CountDualWindow()
	return failure
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingDual) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.resolution) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingDual) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.resolution).Equal(t2.Truncate(c.resolution))
}
