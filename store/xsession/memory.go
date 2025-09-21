//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xcache"
)

var _ Storage = (*MemoryStore)(nil)

func NewMemoryStore(caption int, ttl time.Duration) *MemoryStore {
	return &MemoryStore{
		ttl: ttl,
		db:  xcache.NewLRU[string, *Session](caption),
	}
}

// MemoryStore 在内存中存储 session 信息
type MemoryStore struct {
	ttl time.Duration
	db  *xcache.LRU[string, *Session]
}

func (mem *MemoryStore) Get(ctx context.Context, id string) (*Session, error) {
	return mem.db.Get(ctx, id)
}

func (mem *MemoryStore) GetOrCreate(ctx context.Context, id string) *Session {
	val, err := mem.db.Get(ctx, id)
	if err == nil {
		return val
	}
	return NewValue(id).ToSession(mem)
}

func (mem *MemoryStore) Save(ctx context.Context, session *Session) error {
	return mem.db.Set(ctx, session.ID(), session, mem.ttl)
}
