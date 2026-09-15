//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-15

package xkvmintor

import (
	"context"

	"github.com/xanygo/anygo/xkv"
)

var _ xkv.Storage[any] = (*Monitor[any])(nil)

// Monitor 可以在 KV 操作完成后，执行监控回调
type Monitor[V any] struct {
	// 必填
	Store xkv.Storage[V]

	// 可选，操作完成后调用
	After func(ctx context.Context, dataType DataType, action string, err error, keys ...string)
}

func (m *Monitor[V]) doAfter(ctx context.Context, dataType DataType, action string, err error, keys ...string) {
	if m.After == nil {
		return
	}
	m.After(ctx, dataType, action, err, keys...)
}

func (m *Monitor[V]) String(key string) xkv.String[V] {
	return &monitorString[V]{
		monitor: m,
		key:     key,
		store:   m.Store.String(key),
	}
}

func (m *Monitor[V]) List(key string) xkv.List[V] {
	return &monitorList[V]{
		key:     key,
		store:   m.Store.List(key),
		monitor: m,
	}
}

func (m *Monitor[V]) Hash(key string) xkv.Hash[V] {
	return &monitorHash[V]{
		key:     key,
		monitor: m,
		store:   m.Store.Hash(key),
	}
}

func (m *Monitor[V]) Set(key string) xkv.Set[V] {
	return &monitorSet[V]{
		monitor: m,
		key:     key,
		store:   m.Store.Set(key),
	}
}

func (m *Monitor[V]) ZSet(key string) xkv.ZSet[V] {
	return &monitorZSet[V]{
		store:   m.Store.ZSet(key),
		key:     key,
		monitor: m,
	}
}
