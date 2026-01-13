//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-26

package xcache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xerror"
)

func NewChains[K comparable, V any](caches ...*Chain[K, V]) Cache[K, V] {
	switch len(caches) {
	case 0:
		return &Nop[K, V]{}
	case 1:
		return caches[0].Cache
	default:
		return &chains[K, V]{
			caches: caches,
		}
	}
}

type Chain[K comparable, V any] struct {
	// Cache  必填
	Cache Cache[K, V]

	// DynamicTTL 必填，动态获取数据的缓存有效期,若是 Chains 里的最后一个则不需要
	DynamicTTL func(ctx context.Context, key K, value V) time.Duration

	// WriteTimeout 可选，读取后给未命中缓存的对象，填充缓存的写超时
	WriteTimeout time.Duration
}

func (c *Chain[K, V]) set(ctx context.Context, key K, value V) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()
	ttl := c.DynamicTTL(ctx, key, value)
	_ = c.Cache.Set(ctx, key, value, ttl)
}

func (c *Chain[K, V]) getTimeout() time.Duration {
	if c.WriteTimeout > 0 {
		return c.WriteTimeout
	}
	return 10 * time.Second
}

func (c *Chain[K, V]) CacheTTL(ctx context.Context, key K, value V) time.Duration {
	return c.DynamicTTL(ctx, key, value)
}

func (c *Chain[K, V]) Unwrap() any {
	return c.Cache
}

var _ StringCache = (*chains[string, string])(nil)
var _ HasStats = (*chains[string, string])(nil)

type chains[K comparable, V any] struct {
	caches []*Chain[K, V]
}

func (c *chains[K, V]) Unwrap() []any {
	return xslice.ToAnys(c.caches)
}

func (c *chains[K, V]) Has(ctx context.Context, key K) (has bool, err error) {
	for _, item := range c.caches {
		has, err = item.Cache.Has(ctx, key)
		if has {
			return has, nil
		}
	}
	return false, err
}

func (c *chains[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	var errs []error
	for idx, item := range c.caches {
		value, err := item.Cache.Get(ctx, key)
		if err == nil {
			if idx > 0 {
				go c.setBefore(ctx, idx, key, value)
			}
			return value, nil
		} else if err != nil && !xerror.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return v, errors.Join(errs...)
	}
	return v, xerror.NotFound
}

func (c *chains[K, V]) setBefore(ctx context.Context, idx int, k K, v V) {
	ctx = context.WithoutCancel(ctx)
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	for i := 0; i < idx; i++ {
		c.caches[i].set(ctx, k, v)
	}
}

func (c *chains[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	ctx1 := context.WithoutCancel(ctx)
	// 底层的 cache 一般速度可能会更慢，所以采用异步，并且提前写
	for _, item := range c.caches[1:] {
		go safely.RunCtx(ctx1, func(ctx2 context.Context) {
			ctx3, cancel := context.WithTimeout(ctx2, time.Minute)
			defer cancel()
			_ = item.Cache.Set(ctx3, key, value, ttl)
		})
	}
	err := c.caches[0].Cache.Set(ctx, key, value, ttl)
	return err
}

func (c *chains[K, V]) Delete(ctx context.Context, keys ...K) error {
	var errs []error
	for _, item := range c.caches {
		if err := item.Cache.Delete(ctx, keys...); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

var _ HasStats = (*chains[string, string])(nil)

func (c *chains[K, V]) Stats() Stats {
	for _, item := range c.caches {
		if hs, ok := item.Cache.(HasStats); ok {
			return hs.Stats()
		}
	}
	return Stats{}
}

var _ HasAllStats = (*chains[string, string])(nil)

func (c *chains[K, V]) AllStats() map[string]Stats {
	result := make(map[string]Stats, len(c.caches))
	for idx, item := range c.caches {
		if hs, ok := item.Cache.(HasStats); ok {
			result[fmt.Sprintf("level_%d", idx)] = hs.Stats()
		}
	}
	return result
}
