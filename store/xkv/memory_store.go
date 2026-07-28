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

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// NewMemoryStoreAny 创建一个值类型支持泛型类型的，全内存存储的 KV 存储对象
func NewMemoryStoreAny[V any](coder xcodec.Codec) *Transformer[V] {
	return &Transformer[V]{
		Codec:   coder,
		Storage: NewMemoryStore(),
	}
}

var _ Storage[string] = (*MemoryStore)(nil)

// MemoryStore 底层基础类型为 string 的内存存储实现
type MemoryStore struct {
	base *mem.Base
	once sync.Once
}

func (m *MemoryStore) getBase() *mem.Base {
	m.once.Do(func() {
		m.base = mem.NewBase()
	})
	return m.base
}

func (m *MemoryStore) String(key string) String[string] {
	return &mem.String{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *MemoryStore) List(key string) List[string] {
	return &mem.List{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *MemoryStore) Hash(key string) Hash[string] {
	return &mem.Hash{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *MemoryStore) Set(key string) Set[string] {
	return &mem.Set{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *MemoryStore) ZSet(key string) ZSet[string] {
	return &mem.ZSet{
		Base: m.getBase(),
		Key:  key,
	}
}

func (m *MemoryStore) Delete(ctx context.Context, keys ...string) error {
	return m.getBase().Delete(ctx, keys...)
}

func (m *MemoryStore) Has(ctx context.Context, key string) (found bool, err error) {
	return m.getBase().Has(ctx, key)
}
