//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xerror"
)

type (
	Cache[K comparable, V any] interface {
		// Has 返回，key 是否存在
		//
		//  存在：返回 true,nil
		//  不存在：返回 false,nil
		//  异常： 返回 false,error (error 不会是 xerror.NotFound)
		Has(ctx context.Context, key K) (bool, error)

		// TTL 返回 key 剩余有效期
		//
		//  key 存在：返回 Duration(>0)，nil
		//  key 不存在：返回 0，nil
		//  异常：返回 0，error  (error 不会是 xerror.NotFound)
		TTL(ctx context.Context, key K) (time.Duration, error)

		// Expire 给 key 重新设置有效期
		//
		//  key 存在且有效，返回 nil
		//  key 不存在（或已过期），返回  xerror.NotFound
		//  其他异常，返回 error
		Expire(ctx context.Context, key K, life time.Duration) error

		Getter[K, V]
		Setter[K, V]
		Deleter[K]
	}

	Getter[K comparable, V any] interface {
		// Get 读取数据，
		// error 返回值：
		//  1. 若数据不存在，返回 xerror.NotFound, 可用 IsNotExists 判断
		//  2. 查询到数据，返回 nil
		//  3. 其他异常，返回 error != nil
		Get(ctx context.Context, key K) (value V, err error)
	}

	Setter[K comparable, V any] interface {
		// Set 设置换成，ttl 应 > 0
		Set(ctx context.Context, key K, value V, ttl time.Duration) error
	}

	Deleter[K comparable] interface {
		Delete(ctx context.Context, keys ...K) error
	}
)

type (
	MCache[K comparable, V any] interface {
		Cache[K, V]
		MSetter[K, V]
		MGetter[K, V]
	}

	MSetter[K comparable, V any] interface {
		MSet(ctx context.Context, values map[K]V, ttl time.Duration) error
	}

	MGetter[K comparable, V any] interface {
		// MGet 批量查询，若 key 不存在，则不出现在 result 中
		MGet(ctx context.Context, keys ...K) (result map[K]V, err error)
	}
)

type (
	StringCache Cache[string, string]

	StringMCache MCache[string, string]
)

func IsNotExists(err error) bool {
	return err != nil && xerror.IsNotFound(err)
}

type ValueError[V any] struct {
	Value V     `json:"v,omitempty"`
	Err   error `json:"e,omitempty"`
}

const cacheFileExt = ".cache"

// AsMCache 转换为支持匹配的Cache
//
// concurrency : 并发度，当值 >1 是为并行获取，否则串行处理
func AsMCache[K comparable, V any](c Cache[K, V], concurrency int) MCache[K, V] {
	if mc, ok := c.(MCache[K, V]); ok {
		return mc
	}
	return &toMCache[K, V]{
		Cache:       c,
		concurrency: concurrency,
	}
}

var _ MCache[string, string] = (*toMCache[string, string])(nil)

type toMCache[K comparable, V any] struct {
	Cache[K, V]
	concurrency int // MGet 和 MSet 的并发度，若值 >1 为并发，否则未串行
}

// MGet implements [MCache].
func (t *toMCache[K, V]) MGet(ctx context.Context, keys ...K) (result map[K]V, err error) {
	return mget(ctx, t.Cache, t.concurrency, keys...)
}

// MSet implements [MCache].
func (t *toMCache[K, V]) MSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	return mset(ctx, t.Cache, t.concurrency, values, ttl)
}

func (t *toMCache[K, V]) Unwrap() any {
	return t.Cache
}

func mget[K comparable, V any](ctx context.Context, c Cache[K, V], worker int, keys ...K) (result map[K]V, err error) {
	if len(keys) == 0 {
		return nil, nil
	}
	result = make(map[K]V, len(keys))
	var mux sync.Mutex
	var wg xsync.WaitGroup
	wg.SetLimit(worker)
	for _, key := range keys {
		wg.GoCtxErr(ctx, func(ctx context.Context) error {
			value, err := c.Get(ctx, key)
			if err == nil {
				mux.Lock()
				result[key] = value
				mux.Unlock()
				return nil
			}
			if errors.Is(err, xerror.NotFound) {
				return nil
			}
			return err
		})
	}
	return result, wg.Wait()
}

func mset[K comparable, V any](ctx context.Context, c Cache[K, V], worker int, values map[K]V, ttl time.Duration) (err error) {
	if len(values) == 0 {
		return nil
	}
	var wg xsync.WaitGroup
	wg.SetLimit(worker)
	for key, value := range values {
		wg.GoCtxErr(ctx, func(ctx context.Context) error {
			return c.Set(ctx, key, value, ttl)
		})
	}
	return wg.Wait()
}
