//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcodec"
)

func NewFileStore(dir string, ttl time.Duration) *CacheStore {
	cache := &xcache.File[string, []byte]{
		Dir:   dir,
		Codec: xcodec.Raw,
	}
	return &CacheStore{
		TTL:   ttl,
		Cache: cache,
	}
}

var _ Storage = (*CacheStore)(nil)

// CacheStore 在缓存中存储 session 信息
type CacheStore struct {
	TTL   time.Duration
	Cache xcache.Cache[string, []byte]
}

func (fs *CacheStore) Get(ctx context.Context, id string) (*Session, error) {
	bf, err := fs.Cache.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	val, err := ParserValue(bf)
	if err != nil {
		return nil, err
	}
	val.ID = id
	return val.ToSession(fs), nil
}

func (fs *CacheStore) GetOrCreate(ctx context.Context, id string) *Session {
	se, err := fs.Get(ctx, id)
	if err == nil {
		return se
	}
	val := NewValue(id)
	return val.ToSession(fs)
}

func (fs *CacheStore) Save(ctx context.Context, session *Session) error {
	bf, err := session.Bytes()
	if err != nil {
		return err
	}
	return fs.Cache.Set(ctx, session.ID(), bf, fs.TTL)
}
