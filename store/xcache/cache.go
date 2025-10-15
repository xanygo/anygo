//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"time"

	"github.com/xanygo/anygo/xerror"
)

type (
	Cache[K comparable, V any] interface {
		Getter[K, V]
		Setter[K, V]
		Deleter[K]
	}

	StringCache Cache[string, string]

	Getter[K comparable, V any] interface {
		// Get 读取数据，
		// error 返回值：
		//  1. 若数据不存在，返回 xerror.NotFound, 可用 IsNotExists 判断
		//  2. 查询到数据，返回 nil
		//  3. 其他异常，返回 error != nil
		Get(ctx context.Context, key K) (value V, err error)
	}

	Setter[K comparable, V any] interface {
		Set(ctx context.Context, key K, value V, ttl time.Duration) error
	}

	Deleter[K comparable] interface {
		Delete(ctx context.Context, keys ...K) error
	}

	HasTTL[K comparable, V any] interface {
		CacheTTL(ctx context.Context, key K, value V) time.Duration
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

func IsNotExists(err error) bool {
	return err != nil && xerror.IsNotFound(err)
}

type ValueError[V any] struct {
	Value V     `json:"v,omitempty"`
	Err   error `json:"e,omitempty"`
}

const cacheFileExt = ".cache"
