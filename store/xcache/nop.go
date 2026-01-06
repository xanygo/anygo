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
var _ MCache[string, int] = (*Nop[string, int])(nil)
var _ HasStats = (*Nop[string, int])(nil)

// NopType 不会保存数据的缓存接口实现
//
// 当 Cache 对象的 Nop() 返回 true 时，意味着该 cache 对象在写入时总是成功，读取是总是返回 NotFound 错误
type NopType interface {
	Nop() bool
}

// IsNop 判断是否是一个空的 Logger
func IsNop(c any) bool {
	if c == nil {
		return true
	}
	if nl, ok := c.(NopType); ok && nl.Nop() {
		return true
	}
	return false
}

var _ NopType = (*Nop[string, int])(nil)
var _ Cache[string, int] = (*Nop[string, int])(nil)
var _ MCache[string, int] = (*Nop[string, int])(nil)

// Nop 不会保存任何值的缓存对象
type Nop[K comparable, V any] struct {
	readCnt   atomic.Uint64
	writeCnt  atomic.Uint64
	deleteCnt atomic.Uint64
}

func (n *Nop[K, V]) Has(ctx context.Context, key K) (bool, error) {
	return false, nil
}

func (n *Nop[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	n.readCnt.Add(1)
	return value, xerror.NotFound
}

func (n *Nop[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	n.writeCnt.Add(1)
	return nil
}

func (n *Nop[K, V]) Delete(ctx context.Context, keys ...K) error {
	n.deleteCnt.Add(uint64(len(keys)))
	return nil
}

func (n *Nop[K, V]) MSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	n.writeCnt.Add(uint64(len(values)))
	return nil
}

func (n *Nop[K, V]) MGet(ctx context.Context, keys ...K) (result map[K]V, err error) {
	n.readCnt.Add(uint64(len(keys)))
	return nil, nil
}

func (n *Nop[K, V]) Nop() bool {
	return true
}

func (n *Nop[K, V]) Stats() Stats {
	return Stats{
		Read:   n.readCnt.Load(),
		Write:  n.writeCnt.Load(),
		Delete: n.deleteCnt.Load(),
		Hit:    0,
	}
}
