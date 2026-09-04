//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-26

package xcache

import (
	"context"
	"errors"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
)

var _ Cache[string, string] = (*Wrapper[string, string])(nil)
var _ MCache[string, string] = (*Wrapper[string, string])(nil)

// Wrapper 用于对缓存的 TTL 和 Key 进行动态调整的工具类
//
// 如用 Chains 将多个 cache 组装为一个链式的 Cache 时，可以给不同的 Cache 设置不同的缓存时间
//
//	比如 chainCache.Set（ctx, "key", "value",10 * time.Second）
//	给第一级缓存（纯本地内存缓存），设置 ttl=5s
//	给第二级缓存(远端 redis )，设置为 传入的 10s
type Wrapper[K comparable, V any] struct {
	// Cache 必填，缓存对象
	Cache Cache[K, V]

	// NewLifeFn 可选，依据 key 和 value 动态的 TTL 时间。传入的key是原始的
	//
	// NewLifeFn 和 NewKeyFn 二者至少一个不为空，否则就没有必要使用了
	NewLifeFn func(k K, v V, ttl time.Duration) time.Duration

	// NewKeyFn 可选，用于变换 Key
	NewKeyFn func(k K) K
}

func (w *Wrapper[K, V]) Unwrap() any {
	return w.Cache
}

func (w *Wrapper[K, V]) getDyTTL(key K, value V, ttl time.Duration) time.Duration {
	if w.NewLifeFn != nil {
		return w.NewLifeFn(key, value, ttl)
	}
	return ttl
}

func (w *Wrapper[K, V]) getNewKey(key K) K {
	if w.NewKeyFn != nil {
		return w.NewKeyFn(key)
	}
	return key
}

func (w *Wrapper[K, V]) Has(ctx context.Context, key K) (bool, error) {
	return w.Cache.Has(ctx, w.getNewKey(key))
}

func (w *Wrapper[K, V]) TTL(ctx context.Context, key K) (time.Duration, error) {
	return w.Cache.TTL(ctx, w.getNewKey(key))
}

func (w *Wrapper[K, V]) Expire(ctx context.Context, key K, ttl time.Duration) error {
	return w.Cache.Expire(ctx, w.getNewKey(key), ttl)
}

func (w *Wrapper[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	return w.Cache.Get(ctx, w.getNewKey(key))
}

func (w *Wrapper[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	return w.Cache.Set(ctx, w.getNewKey(key), value, w.getDyTTL(key, value, ttl))
}

func (w *Wrapper[K, V]) Delete(ctx context.Context, keys ...K) error {
	if len(keys) == 0 {
		return nil
	}
	for i := 0; i < len(keys); i++ {
		keys[i] = w.getNewKey(keys[i])
	}
	return w.Cache.Delete(ctx, keys...)
}

func (w *Wrapper[K, V]) MSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	if w.NewLifeFn == nil {
		if w.NewKeyFn == nil {
			return w.doMSet(ctx, values, ttl)
		}
		kvs := make(map[K]V, len(values))
		for k, v := range values {
			kvs[w.NewKeyFn(k)] = v
		}
		return w.doMSet(ctx, kvs, ttl)
	}
	chunks := make(map[time.Duration]map[K]V, 0)
	var num int
	for k, v := range values {
		nt := w.NewLifeFn(k, v, ttl)
		if _, has := chunks[nt]; !has {
			chunks[nt] = make(map[K]V, len(values)-num)
		}
		if w.NewKeyFn == nil {
			chunks[nt][k] = v
		} else {
			chunks[nt][w.NewKeyFn(k)] = v
		}
		num++
	}

	if len(chunks) == 1 {
		for nt, kvs := range chunks {
			return w.doMSet(ctx, kvs, nt)
		}
	} else {
		var wg xsync.WaitGroup
		for nt, kvs := range chunks {
			wg.GoCtxErr(ctx, func(ctx context.Context) error {
				return w.doMSet(ctx, kvs, nt)
			})
		}
		return wg.Wait()
	}

	// 理论上不会触达
	return errors.New("unreachable code")
}

func (w *Wrapper[K, V]) doMSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	if mc, ok := w.Cache.(MCache[K, V]); ok {
		return mc.MSet(ctx, values, ttl)
	}
	return mset(ctx, w.Cache, -1, values, ttl)
}

func (w *Wrapper[K, V]) MGet(ctx context.Context, keys ...K) (result map[K]V, err error) {
	if len(keys) == 0 {
		return nil, nil
	}
	if w.NewKeyFn == nil {
		return mget(ctx, w.Cache, -1, keys...)
	}

	mapping := make(map[K]K, len(keys))
	newKeys := make([]K, len(keys))
	for i, oldKey := range keys {
		newKey := w.NewKeyFn(oldKey)
		newKeys[i] = newKey
		mapping[newKey] = oldKey
	}
	values, err := mget(ctx, w.Cache, -1, newKeys...)
	if len(values) > 0 {
		result = make(map[K]V, len(values))
		for k, v := range values {
			result[mapping[k]] = v
		}
	}
	return result, err
}

var _ HasStats = (*Wrapper[string, string])(nil)

func (w *Wrapper[K, V]) Stats() Stats {
	if hs, ok := w.Cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{}
}

var _ HasAllStats = (*Wrapper[string, string])(nil)

func (w *Wrapper[K, V]) AllStats() map[string]Stats {
	if hs, ok := w.Cache.(HasAllStats); ok {
		return hs.AllStats()
	}
	return nil
}
