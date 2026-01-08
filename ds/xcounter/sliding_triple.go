//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-07

package xcounter

import (
	"sync"
	"time"
)

// NewSlidingWindowTriple 创建一个滑动窗口计数器
func NewSlidingWindowTriple(windowSize, bucketSize time.Duration) *SlidingWindowTriple {
	if bucketSize <= 0 {
		bucketSize = time.Second
	}
	if windowSize < bucketSize {
		windowSize = bucketSize
	}
	size := int(windowSize / bucketSize)
	return &SlidingWindowTriple{
		bucketSize: bucketSize,
		windowSize: windowSize,
		buckets:    make([]bucket3, size),
		start:      time.Now(),
	}
}

// SlidingWindowTriple 可同时记录 成功数、失败数、耗时的滑动窗口计数器
type SlidingWindowTriple struct {
	mu         sync.Mutex
	bucketSize time.Duration // 每个桶的时间粒度
	windowSize time.Duration // 总窗口时长
	buckets    []bucket3
	start      time.Time

	totalSuccess int64
	totalFailure int64
	totalCost    time.Duration
}

func (c *SlidingWindowTriple) WindowSize() time.Duration {
	return c.windowSize
}

func (c *SlidingWindowTriple) BucketSize() time.Duration {
	return c.bucketSize
}

type bucket3 struct {
	ts      time.Time
	success int64
	failure int64
	cost    time.Duration
}

func (c *SlidingWindowTriple) IncrAuto(err error, cost time.Duration) {
	if err == nil {
		c.IncrN(1, 0, cost)
	} else {
		c.IncrN(0, 1, cost)
	}
}

// IncrN 往计数器加 N （ N = success + failure ），cost 是 N 的总耗时
// 若是 N=1 则 success + failure =1， cost 是这一次的耗时
func (c *SlidingWindowTriple) IncrN(success, failure int64, cost time.Duration) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalSuccess += success
	c.totalFailure += failure
	c.totalCost += cost

	index := c.indexFor(now)

	if !c.sameSlot(c.buckets[index].ts, now) {
		// 重置过期的桶
		c.buckets[index].ts = now.Truncate(c.bucketSize)
		c.buckets[index].success = 0
		c.buckets[index].failure = 0
		c.buckets[index].cost = 0
	}
	c.buckets[index].success += success
	c.buckets[index].failure += failure
	c.buckets[index].cost += cost
}

func (c *SlidingWindowTriple) LifetimeCounts() (success, failure int64, cost time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSuccess, c.totalFailure, c.totalCost
}

func (c *SlidingWindowTriple) LifetimeTotal() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSuccess + c.totalFailure
}

func (c *SlidingWindowTriple) LifetimeCost() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalCost
}

func (c *SlidingWindowTriple) LifetimeSuccess() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalSuccess
}

func (c *SlidingWindowTriple) LifetimeFailure() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalFailure
}

// 当前窗口统计

// WindowCounts 返回整个窗口的计数
func (c *SlidingWindowTriple) WindowCounts() (success, failure int64, cost time.Duration) {
	return c.Counts(c.windowSize)
}

// WindowTotal 返回整个窗口的计数: 成功数 + 失败数 和 耗时
func (c *SlidingWindowTriple) WindowTotal() (int64, time.Duration) {
	success, failure, cost := c.WindowCounts()
	return success + failure, cost
}

// WindowSuccess 返回整个窗口的成功数
func (c *SlidingWindowTriple) WindowSuccess() int64 {
	success, _, _ := c.WindowCounts()
	return success
}

// WindowFailure 返回整个窗口的失败数
func (c *SlidingWindowTriple) WindowFailure() int64 {
	_, failure, _ := c.WindowCounts()
	return failure
}

func (c *SlidingWindowTriple) WindowCost() time.Duration {
	_, _, cost := c.WindowCounts()
	return cost
}

// 任意 duration 内统计

// Counts 返回指定时间返回内的计数
func (c *SlidingWindowTriple) Counts(d time.Duration) (success, failure int64, cost time.Duration) {
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
			cost += b.cost
		}
	}
	return success, failure, cost
}

// Success 返回指定时间返回内的成功总计数
func (c *SlidingWindowTriple) Success(d time.Duration) int64 {
	success, _, _ := c.Counts(d)
	return success
}

// Failure 返回指定时间返回内的失败总计数
func (c *SlidingWindowTriple) Failure(d time.Duration) int64 {
	_, failure, _ := c.Counts(d)
	return failure
}

// Count 获取时间段内的 成功数 + 失败数
func (c *SlidingWindowTriple) Count(d time.Duration) int64 {
	success, failure, _ := c.Counts(d)
	return success + failure
}

func (c *SlidingWindowTriple) Cost(d time.Duration) time.Duration {
	_, _, cost := c.Counts(d)
	return cost
}

// 最近 N 个窗口统计

func (c *SlidingWindowTriple) WindowNCounts(n int) (success, failure int64, cost time.Duration) {
	return c.Counts(c.bucketSize * time.Duration(n))
}

func (c *SlidingWindowTriple) WindowNTotal(n int) (int64, time.Duration) {
	s, f, cost := c.WindowNCounts(n)
	return s + f, cost
}

func (c *SlidingWindowTriple) WindowNSuccess(n int) int64 {
	success, _, _ := c.WindowNCounts(n)
	return success
}

func (c *SlidingWindowTriple) WindowNFailure(n int) int64 {
	_, failure, _ := c.WindowNCounts(n)
	return failure
}

func (c *SlidingWindowTriple) WindowNCost(n int) time.Duration {
	_, _, cost := c.WindowNCounts(n)
	return cost
}

// indexFor 计算当前时间对应的桶索引
func (c *SlidingWindowTriple) indexFor(t time.Time) int {
	elapsed := t.Sub(c.start)
	return int(elapsed/c.bucketSize) % len(c.buckets)
}

// sameSlot 判断两个时间是否落在同一个桶内
func (c *SlidingWindowTriple) sameSlot(t1, t2 time.Time) bool {
	return t1.Truncate(c.bucketSize).Equal(t2.Truncate(c.bucketSize))
}

// Export 导出统计数据
func (c *SlidingWindowTriple) Export(ts ...time.Duration) map[string]any {
	result := make(map[string]any, len(ts)+3)
	result["Meta"] = map[string]any{
		"WindowSize": c.windowSize.String(),
		"BucketSize": c.bucketSize.String(),
		"Start":      c.start.Format(time.DateTime),
	}

	success, fail, cost := c.LifetimeCounts()
	result["Lifetime"] = map[string]any{
		"Success": success,
		"failure": fail,
		"CostAvg": c.costAvg(success+fail, cost).String(),
	}

	success, fail, cost = c.WindowCounts()
	result["Window"] = map[string]any{
		"Success": success,
		"failure": fail,
		"CostAvg": c.costAvg(success+fail, cost).String(),
	}

	for _, t := range ts {
		if t >= c.windowSize {
			continue
		}
		success, fail, cost = c.Counts(t)
		if success == 0 && fail == 0 && cost == 0 {
			continue
		}
		result[t.String()] = map[string]any{
			"Success": success,
			"failure": fail,
			"CostAvg": c.costAvg(success+fail, cost).String(),
		}
	}
	return result
}

func (c *SlidingWindowTriple) costAvg(num int64, cost time.Duration) time.Duration {
	if num == 0 {
		return 0
	}
	return cost / time.Duration(num)
}
