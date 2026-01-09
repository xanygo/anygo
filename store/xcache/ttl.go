//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-26

package xcache

import (
	"context"
	"time"
)

var _ HasTTL[string, string] = (*TTLWrapper[string, string])(nil)
var _ Cache[string, string] = (*TTLWrapper[string, string])(nil)
var _ NopCache = (*TTLWrapper[string, string])(nil)

// TTLWrapper 用于对缓存的 TTL 进行动态调整的工具类
//
// 如用 Chains 将多个 cache 组装为一个链式的 Cache 时，可以给不同的 Cache 设置不同的缓存时间
//
//	比如 chainCache.Set（ctx, "key", "value",10 * time.Second）
//	给第一级缓存（纯本地内存缓存），设置 ttl=5s
//	给第二级缓存(远端 redis )，设置为 传入的 10s
type TTLWrapper[K comparable, V any] struct {
	// Cache 必填，缓存对象
	Cache Cache[K, V]

	// Fixed 固定的 TTL 时间，Fixed 和 Dynamic 二选一
	Fixed time.Duration

	// Dynamic 动态的 TTL 时间
	Dynamic func(ctx context.Context, k K, v V, ttl time.Duration) time.Duration
}

func (tw *TTLWrapper[K, V]) Nop() bool {
	return IsNop(tw.Cache)
}

func (tw *TTLWrapper[K, V]) CacheTTL(ctx context.Context, key K, value V) time.Duration {
	if tw.Dynamic != nil {
		return tw.Dynamic(ctx, key, value, tw.Fixed)
	}
	return tw.Fixed
}

func (tw *TTLWrapper[K, V]) Has(ctx context.Context, key K) (bool, error) {
	return tw.Cache.Has(ctx, key)
}

func (tw *TTLWrapper[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	return tw.Cache.Get(ctx, key)
}

func (tw *TTLWrapper[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	if tw.Dynamic != nil {
		ttl = tw.Dynamic(ctx, key, value, ttl)
	} else if tw.Fixed > 0 {
		ttl = tw.Fixed
	}
	return tw.Cache.Set(ctx, key, value, ttl)
}

func (tw *TTLWrapper[K, V]) Delete(ctx context.Context, keys ...K) error {
	return tw.Cache.Delete(ctx, keys...)
}

var _ HasStats = (*TTLWrapper[string, string])(nil)

func (tw *TTLWrapper[K, V]) Stats() Stats {
	if hs, ok := tw.Cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{}
}

var _ HasAllStats = (*TTLWrapper[string, string])(nil)

func (tw *TTLWrapper[K, V]) AllStats() map[string]Stats {
	if hs, ok := tw.Cache.(HasAllStats); ok {
		return hs.AllStats()
	}
	return nil
}
