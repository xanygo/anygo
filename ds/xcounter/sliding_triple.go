//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-07

package xcounter

import (
	"sync"
	"time"
)

// NewSlidingTriple 创建一个滑动窗口计数器
func NewSlidingTriple(window, resolution time.Duration) *SlidingTriple {
	if resolution <= 0 {
		resolution = time.Second
	}
	if window < resolution {
		window = resolution
	}
	size := int(window / resolution)
	return &SlidingTriple{
		resolution: resolution,
		window:     window,
		buckets:    make([]bucket3, size),
		start:      time.Now(),
	}
}

// SlidingTriple 可同时记录 成功数、失败数、耗时的滑动窗口计数器
type SlidingTriple struct {
	mu         sync.Mutex
	resolution time.Duration // 每个桶的时间粒度
	window     time.Duration // 总窗口时长
	buckets    []bucket3
	start      time.Time
}

func (c *SlidingTriple) Window() time.Duration {
	return c.window
}

func (c *SlidingTriple) Resolution() time.Duration {
	return c.resolution
}

type bucket3 struct {
	ts      time.Time
	success int64
	failure int64
	cost    time.Duration
}

func (c *SlidingTriple) IncrAuto(err error, cost time.Duration) {
	if err == nil {
		c.IncrN(1, 0, cost)
	} else {
		c.IncrN(0, 1, cost)
	}
}

// IncrN 往计数器加 N （ N = success + failure ），cost 是 N 的总耗时
// 若是 N=1 则 success + failure =1， cost 是这一次的耗时
func (c *SlidingTriple) IncrN(success, failure int64, cost time.Duration) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	index := c.indexFor(now)

	if !c.sameSlot(c.buckets[index].ts, now) {
		// 重置过期的桶
		c.buckets[index].ts = now.Truncate(c.resolution)
		c.buckets[index].success = 0
		c.buckets[index].failure = 0
		c.buckets[index].cost = 0
	}
	c.buckets[index].success += success
	c.buckets[index].failure += failure
	c.buckets[index].cost += cost
}

// TotalTriple 返回整个窗口的计数
func (c *SlidingTriple) TotalTriple() (success, failure int64, cost time.Duration) {
	return c.CountTriple(c.window)
}

// Total 返回整个窗口的计数: 成功数 + 失败数 和 耗时
func (c *SlidingTriple) Total() (int64, time.Duration) {
	success, failure, cost := c.TotalTriple()
	return success + failure, cost
}

// TotalSuccess 返回整个窗口的成功数
func (c *SlidingTriple) TotalSuccess() int64 {
	success, _, _ := c.TotalTriple()
	return success
}

// TotalFailure 返回整个窗口的失败数
func (c *SlidingTriple) TotalFailure() int64 {
	_, failure, _ := c.TotalTriple()
	return failure
}

// CountSuccess 返回指定时间返回内的成功总计数
func (c *SlidingTriple) CountSuccess(d time.Duration) int64 {
	success, _, _ := c.CountTriple(d)
	return success
}

// CountFailure 返回指定时间返回内的失败总计数
func (c *SlidingTriple) CountFailure(d time.Duration) int64 {
	_, failure, _ := c.CountTriple(d)
	return failure
}

// CountTriple 返回指定时间返回内的计数
func (c *SlidingTriple) CountTriple(d time.Duration) (success, failure int64, cost time.Duration) {
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
			cost += b.cost
		}
	}
	return success, failure, cost
}

// Count 获取时间段内的 成功数 + 失败数
func (c *SlidingTriple) Count(d time.Duration) int64 {
	success, failure, _ := c.CountTriple(d)
	return success + failure
}

func (c *SlidingTriple) Cost(d time.Duration) time.Duration {
	_, _, cost := c.CountTriple(d)
	return cost
}

func (c *SlidingTriple) CountTripleWindowN(n int) (success, failure int64, cost time.Duration) {
	return c.CountTriple(c.window * time.Duration(n))
}

func (c *SlidingTriple) CountSuccessWindowN(n int) int64 {
	success, _, _ := c.CountTripleWindowN(n)
	return success
}

func (c *SlidingTriple) CountFailureWindowN(n int) int64 {
	_, failure, _ := c.CountTripleWindowN(n)
	return failure
}

func (c *SlidingTriple) CountWindowN(n int) int64 {
	success, failure, _ := c.CountTripleWindowN(n)
	return success + failure
}

// CountTripleWindow 返回最近一个窗口周期的总计数
func (c *SlidingTriple) CountTripleWindow() (success, failure int64, cost time.Duration) {
	return c.CountTripleWindowN(1)
}

// CountWindow 返回最近一个窗口周期的总计数（成功+失败）
func (c *SlidingTriple) CountWindow() int64 {
	success, failure, _ := c.CountTripleWindow()
	return success + failure
}

// CostWindow 返回最近一个窗口周期的总耗时
func (c *SlidingTriple) CostWindow() time.Duration {
	_, _, cost := c.CountTripleWindow()
	return cost
}

// CountSuccessWindow 返回最近一个窗口周期的成功总次数
func (c *SlidingTriple) CountSuccessWindow() int64 {
	success, _, _ := c.CountTripleWindow()
	return success
}

// CountFailureWindow 返回最近一个窗口周期的失败总次数
func (c *SlidingTriple) CountFailureWindow() int64 {
	_, failure, _ := c.CountTripleWindow()
	return failure
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingTriple) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.resolution) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingTriple) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.resolution).Equal(t2.Truncate(c.resolution))
}

// Export 到处统计数据
func (c *SlidingTriple) Export(ts ...time.Duration) map[string]any {
	result := make(map[string]any, len(ts)+3)
	result["Window"] = c.window.String()
	result["Resolution"] = c.resolution.String()

	success, fail, cost := c.TotalTriple()
	result["All"] = map[string]any{
		"Success": success,
		"Fail":    fail,
		"CostAvg": c.costAvg(success+fail, cost).String(),
	}
	for _, t := range ts {
		if c.window <= t {
			continue
		}
		success, fail, cost = c.CountTriple(t)
		if success == 0 && fail == 0 && cost == 0 {
			continue
		}
		result[t.String()] = map[string]any{
			"Success": success,
			"Fail":    fail,
			"CostAvg": c.costAvg(success+fail, cost).String(),
		}
	}
	return result
}

func (c *SlidingTriple) costAvg(num int64, cost time.Duration) time.Duration {
	if num == 0 {
		return 0
	}
	return cost / time.Duration(num)
}
