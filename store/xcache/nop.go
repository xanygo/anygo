//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcache

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/xerror"
)

var _ Cache[string, int] = (*Nop[string, int])(nil)
var _ HasStats = (*Nop[string, int])(nil)

// Nop 不会保存任何值的缓存对象
type Nop[K comparable, V any] struct {
	getCnt    atomic.Uint64 // 调用 Get 方法的次数
	setCnt    atomic.Uint64 // 调用 Set 方法的次数
	deleteCnt atomic.Uint64 // 调用 Delete 方法的次数
}

func (n *Nop[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	n.getCnt.Add(1)
	return value, xerror.NotFound
}

func (n *Nop[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	n.setCnt.Add(1)
	return nil
}

func (n *Nop[K, V]) Delete(ctx context.Context, keys ...K) error {
	n.deleteCnt.Add(1)
	return nil
}

func (n *Nop[K, V]) Stats() Stats {
	return Stats{
		Get:    n.getCnt.Load(),
		Set:    n.setCnt.Load(),
		Delete: n.deleteCnt.Load(),
		Hit:    0,
		Keys:   0,
	}
}
