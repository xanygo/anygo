//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-23

package xkv

import (
	"context"
	"sync"

	"github.com/xanygo/anygo/store/xkv/internal/mem"
	"github.com/xanygo/anygo/xcodec"
)

func NewMemory() *Memory {
	return &Memory{}
}

// NewMemoryAny 创建一个值类型支持泛型类型的，全内存存储的 KV 存储对象
func NewMemoryAny[V any](coder xcodec.Codec) *Transformer[V] {
	return &Transformer[V]{
		Codec:   coder,
		Storage: NewMemory(),
	}
}

var _ Storage[string] = (*Memory)(nil)

// Memory 底层基础类型为 string 的内存存储实现
type Memory struct {
	base *mem.Base
	once sync.Once
}

func (m *Memory) getBase() *mem.Base {
	m.once.Do(func() {
		m.base = mem.NewBase()
	})
	return m.base
}

func (m *Memory) String(key string) String[string] {
	return &mem.String{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *Memory) List(key string) List[string] {
	return &mem.List{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *Memory) Hash(key string) Hash[string] {
	return &mem.Hash{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *Memory) Set(key string) Set[string] {
	return &mem.Set{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *Memory) ZSet(key string) ZSet[string] {
	return &mem.ZSet{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *Memory) Delete(ctx context.Context, keys ...string) error {
	return m.getBase().Delete(ctx, keys...)
}

func (m *Memory) Has(ctx context.Context, key string) (found bool, err error) {
	return m.getBase().Has(ctx, key)
}
