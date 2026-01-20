//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-28

package xcache

import (
	"context"
	"time"

	"github.com/xanygo/anygo/ds/xcounter"
	"github.com/xanygo/anygo/internal/ztypes"
	"github.com/xanygo/anygo/xcodec"
)

type (
	HasStats interface {
		Stats() Stats
	}

	// HasAllStats chains 这类包含有多个 cache 的场景使用
	HasAllStats interface {
		AllStats() map[string]Stats
	}
)

// Stats 统计信息
type Stats struct {
	Capacity int    // 最多缓存 keys 个数（近似，实际可能会超出），部分 cache 有
	Keys     int64  // 个数
	Read     uint64 // 读取的 key 的 总数量，包括 GET 和 MGet
	Write    uint64
	Delete   uint64
	Hit      uint64
}

func (s Stats) String() string {
	str, _ := xcodec.EncodeToString(xcodec.JSON, s)
	return str
}

// HitRate 命中率
func (s Stats) HitRate() float64 {
	if s.Read == 0 {
		return 0
	}
	hit := float64(s.Hit) / float64(s.Read)
	return hit
}

// GetStats 读取缓存对象的 统计信息
func GetStats(cache any) Stats {
	if hs, ok := cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{}
}

type StatsRegistry ztypes.Registry[string, HasStats]

var statsRegistry StatsRegistry = ztypes.NewRegistry[string, HasStats]()

func Registry() StatsRegistry {
	return statsRegistry
}

func NewLatencyObserver[K comparable, V any](cache Cache[K, V], window, resolution time.Duration) *LatencyObserver[K, V] {
	return &LatencyObserver[K, V]{
		next: cache,
		get:  xcounter.NewSlidingWindowTriple(window, resolution),
		set:  xcounter.NewSlidingWindowTriple(window, resolution),
		del:  xcounter.NewSlidingWindowTriple(window, resolution),
	}
}

var _ Cache[string, any] = (*LatencyObserver[string, any])(nil)
var _ HasStats = (*LatencyObserver[string, any])(nil)

// LatencyObserver 封装，以统计周期范围内各项指标的执行次数和耗时
type LatencyObserver[K comparable, V any] struct {
	next Cache[K, V]
	get  *xcounter.SlidingWindowTriple
	set  *xcounter.SlidingWindowTriple
	del  *xcounter.SlidingWindowTriple
}

func (lo *LatencyObserver[K, V]) Has(ctx context.Context, key K) (bool, error) {
	return lo.next.Has(ctx, key)
}

func (lo *LatencyObserver[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	start := time.Now()
	value, err = lo.next.Get(ctx, key)
	cost := time.Since(start)
	lo.get.IncrAuto(err, cost)
	return value, err
}

func (lo *LatencyObserver[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	start := time.Now()
	err := lo.next.Set(ctx, key, value, ttl)
	cost := time.Since(start)
	lo.set.IncrAuto(err, cost)
	return err
}

func (lo *LatencyObserver[K, V]) Delete(ctx context.Context, keys ...K) error {
	start := time.Now()
	err := lo.next.Delete(ctx, keys...)
	cost := time.Since(start)
	lo.del.IncrAuto(err, cost)
	return err
}

func (lo *LatencyObserver[K, V]) Stats() Stats {
	if s, ok := lo.next.(HasStats); ok {
		return s.Stats()
	}
	return Stats{}
}

func (lo *LatencyObserver[K, V]) Unwrap() any {
	return lo.next
}

type HasLatencyStats interface {
	LatencyStats() map[string]any
}

func (lo *LatencyObserver[K, V]) LatencyStats() map[string]any {
	result := map[string]any{
		"Get":    lo.get.Export(time.Hour, 10*time.Minute, 5*time.Minute, time.Minute),
		"Set":    lo.set.Export(time.Hour, 10*time.Minute, 5*time.Minute, time.Minute),
		"Delete": lo.del.Export(time.Hour, 10*time.Minute, 5*time.Minute, time.Minute),
	}
	return result
}
