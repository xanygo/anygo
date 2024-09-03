//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcache

import (
	"context"
	"time"
)

var _ Cache[string, int] = (*Nop[string, int])(nil)

// Nop 不会保存任何值的缓存对象
type Nop[K comparable, V any] struct{}

func (n *Nop[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	return value, ErrNil
}

func (n *Nop[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	return nil
}

func (n *Nop[K, V]) Delete(ctx context.Context, keys ...K) error {
	return nil
}
