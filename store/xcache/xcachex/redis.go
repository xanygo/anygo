//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xcachex

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/xerror"
)

var _ xcache.StringCache = (*Redis)(nil)

type Redis struct {
	KeyPrefix string
	Client    *xredis.Client

	getCnt    atomic.Uint64 // 调用 Get 方法的次数
	setCnt    atomic.Uint64 // 调用 Set 方法的次数
	deleteCnt atomic.Uint64 // 调用 Delete 方法的次数
	hitCnt    atomic.Uint64 // 调用 Get 命中缓存的次数
}

func (r *Redis) Get(ctx context.Context, key string) (value string, err error) {
	r.getCnt.Add(1)
	value, err = r.Client.Get(ctx, r.KeyPrefix+key)
	if err == nil {
		r.hitCnt.Add(1)
		return value, nil
	}
	if errors.Is(err, xredis.ErrNil) {
		return value, xerror.NotFound
	}
	return value, err
}

func (r *Redis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	r.setCnt.Add(1)
	return r.Client.Set(ctx, r.KeyPrefix+key, value, ttl)
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	r.deleteCnt.Add(1)
	keysNew := make([]string, len(keys))
	for i, key := range keys {
		keysNew[i] = r.KeyPrefix + key
	}
	_, err := r.Client.Del(ctx, keysNew...)
	return err
}

func (r *Redis) Stats() xcache.Stats {
	return xcache.Stats{
		Get:    r.getCnt.Load(),
		Set:    r.setCnt.Load(),
		Delete: r.deleteCnt.Load(),
		Hit:    r.hitCnt.Load(),
		Keys:   -1,
	}
}
