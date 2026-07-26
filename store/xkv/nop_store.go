//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package xkv

import (
	"context"

	"github.com/xanygo/anygo/store/xkv/internal/nop"
)

var _ Storage[any] = (*NopStorage[any])(nil)

// NopStorage 一个黑洞存储实现
type NopStorage[V any] struct{}

func (n NopStorage[V]) String(key string) String[V] {
	return nop.String[V]{}
}

func (n NopStorage[V]) List(key string) List[V] {
	return nop.List[V]{}
}

func (n NopStorage[V]) Hash(key string) Hash[V] {
	return nop.Hash[V]{}
}

func (n NopStorage[V]) Set(key string) Set[V] {
	return nop.Set[V]{}
}

func (n NopStorage[V]) ZSet(key string) ZSet[V] {
	return nop.ZSet[V]{}
}

func (n NopStorage[V]) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n NopStorage[V]) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
