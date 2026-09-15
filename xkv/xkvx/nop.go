//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package xkvx

import (
	"context"

	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xkv/internal/nop"
)

var _ xkv.Storage[any] = (*Nop[any])(nil)

// Nop 一个黑洞存储实现
type Nop[V any] struct{}

func (n Nop[V]) String(key string) xkv.String[V] {
	return nop.String[V]{}
}

func (n Nop[V]) List(key string) xkv.List[V] {
	return nop.List[V]{}
}

func (n Nop[V]) Hash(key string) xkv.Hash[V] {
	return nop.Hash[V]{}
}

func (n Nop[V]) Set(key string) xkv.Set[V] {
	return nop.Set[V]{}
}

func (n Nop[V]) ZSet(key string) xkv.ZSet[V] {
	return nop.ZSet[V]{}
}

func (n Nop[V]) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n Nop[V]) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
