//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"time"
)

var _ Cache[string, string] = (*Reader[string, string])(nil)

// Reader 自动读取数据并设置缓存
type Reader[K comparable, V any] struct {
	// New 创建新值的函数,必填
	New func(ctx context.Context, key K) (V, error)

	// Cache 缓存对象，必填
	Cache Cache[K, ValueError[V]]

	// TTL 缓存有效期，必填
	TTL time.Duration

	// FailTTL 当 New 方法创建对象失败的时候，可选，缓存的有效期，默认为 0。
	// > 0 时生效存储 New 失败的 error 信息
	FailTTL time.Duration
}

func (rd *Reader[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	val := ValueError[V]{
		Value: value,
	}
	return rd.Cache.Set(ctx, key, val, ttl)
}

// Read 读取缓存中的值
func (rd *Reader[K, V]) Read(ctx context.Context, key K) (v V, err error) {
	value, err := rd.Cache.Get(ctx, key)
	if err == nil {
		return value.Value, value.Err
	}
	return v, err
}

// Get 读取数据，若没有，会先查询
func (rd *Reader[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	value, err := rd.Cache.Get(ctx, key)
	if err == nil {
		return value.Value, value.Err
	}
	if !IsNotExists(err) {
		return v, err
	}
	v, err = rd.New(ctx, key)

	value = ValueError[V]{
		Value: v,
		Err:   err,
	}
	if err == nil {
		err = rd.Cache.Set(ctx, key, value, rd.TTL)
	} else if rd.FailTTL > 0 {
		rd.Cache.Set(ctx, key, value, rd.FailTTL)
	}
	return v, err
}

// Flush 刷新缓存的数据
func (rd *Reader[K, V]) Flush(ctx context.Context, key K) (v V, err error) {
	v, err = rd.New(ctx, key)
	if err != nil {
		return v, err
	}
	value := ValueError[V]{
		Value: v,
		Err:   err,
	}
	err = rd.Cache.Set(ctx, key, value, rd.TTL)
	return v, err
}

func (rd *Reader[K, V]) Delete(ctx context.Context, keys ...K) error {
	return rd.Cache.Delete(ctx, keys...)
}
