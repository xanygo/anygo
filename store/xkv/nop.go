//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package xkv

import (
	"context"

	"github.com/xanygo/anygo/store/xkv/internal/nop"
)

var _ Storage[any] = (*Nop[any])(nil)

// Nop 一个黑洞存储实现
type Nop[V any] struct{}

func (n Nop[V]) String(key string) String[V] {
	return nop.String[V]{}
}

func (n Nop[V]) List(key string) List[V] {
	return nop.List[V]{}
}

func (n Nop[V]) Hash(key string) Hash[V] {
	return nop.Hash[V]{}
}

func (n Nop[V]) Set(key string) Set[V] {
	return nop.Set[V]{}
}

func (n Nop[V]) ZSet(key string) ZSet[V] {
	return nop.ZSet[V]{}
}

func (n Nop[V]) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n Nop[V]) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
