//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-31

package zcache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Map 一个简单的，使用 sync.Map 作为存储的缓存
type Map[K any, V any] struct {
	// New 创建新值的函数,必填
	New func(ctx context.Context, key K) (V, error)

	// TTL 缓存有效期，可选，当为 0 时，默认值为 1 分钟
	TTL time.Duration

	// FailTTL 当 New 方法创建对象失败的时候，可选，缓存的有效期，默认为 0。
	// > 0 时生效存储 New 失败的 error 信息
	FailTTL time.Duration

	// Caption 容量，可选，当为 0时，默认值为 100000
	// 这个值是一个近似值
	Caption int64

	count  atomic.Int64
	values sync.Map
}

func (mc *Map[K, V]) getCaption() int64 {
	if mc.Caption == 0 {
		return 100000
	}
	return mc.Caption
}

func (mc *Map[K, V]) getTTL() time.Duration {
	if mc.TTL == 0 {
		return time.Minute
	}
	return mc.TTL
}

// Get 读取一个值
func (mc *Map[K, V]) Get(ctx context.Context, key K) (V, error) {
	cv, has := mc.values.Load(key)
	if has {
		if vv := cv.(*value[V]); vv.IsOK() {
			return vv.payload, vv.err
		}
	}
	nv, err := mc.New(ctx, key)
	if err == nil {
		mc.store(key, nv, nil, mc.getTTL())
	} else {
		if mc.FailTTL > 0 {
			mc.store(key, nv, err, mc.FailTTL)
		}
	}
	return nv, err
}

func (mc *Map[K, V]) store(key K, nv V, err error, ttl time.Duration) {
	cv := &value[V]{
		payload: nv,
		err:     err,
		expired: time.Now().Add(ttl),
	}
	_, hasOld := mc.values.LoadOrStore(key, cv)
	if hasOld {
		return
	}
	num := mc.count.Add(1)
	caption := mc.getCaption()
	if del := num - caption; del > caption/4 {
		mc.clear(key, int(del))
	}
}

func (mc *Map[K, V]) clear(notKey any, needDel int) {
	delKeys := make([]any, 0, 5)
	var loop int
	mc.values.Range(func(k, v any) bool {
		if k == notKey {
			return true
		}
		loop++

		// 当超出 Caption 10个以上的时候，直接删除一些
		if needDel > 10 && loop < 5 {
			delKeys = append(delKeys, k)
			return true
		}

		if loop >= 5 {
			// 当查找了几次没有找到过期数据的时候，直接删除一项
			delKeys = append(delKeys, k)
			return false
		}

		cv := v.(*value[V])
		if cv.IsOK() {
			return true
		}
		delKeys = append(delKeys, k)
		return false
	})

	for i := 0; i < len(delKeys); i++ {
		mc.Delete(delKeys[i].(K))
	}
}

// Delete 删除值
func (mc *Map[K, V]) Delete(key K) int {
	_, ok := mc.values.LoadAndDelete(key)
	if ok {
		mc.count.Add(-1)
		return 1
	}
	return 0
}

type value[V any] struct {
	expired time.Time
	payload V
	err     error
}

func (v *value[V]) IsOK() bool {
	return time.Now().Before(v.expired)
}