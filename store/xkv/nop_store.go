//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package xkv

import (
	"context"

	"github.com/xanygo/anygo/store/xkv/internal/nop"
)

var _ Storage[any] = (*NopStore[any])(nil)

// NopStore 一个黑洞存储实现
type NopStore[V any] struct{}

func (n NopStore[V]) String(key string) String[V] {
	return nop.String[V]{}
}

func (n NopStore[V]) List(key string) List[V] {
	return nop.List[V]{}
}

func (n NopStore[V]) Hash(key string) Hash[V] {
	return nop.Hash[V]{}
}

func (n NopStore[V]) Set(key string) Set[V] {
	return nop.Set[V]{}
}

func (n NopStore[V]) ZSet(key string) ZSet[V] {
	return nop.ZSet[V]{}
}

func (n NopStore[V]) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n NopStore[V]) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
