//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xsession

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
)

func NewFileStore(dir string, ttl time.Duration) *CacheStore {
	cache := &xcache.File[string, string]{
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
	Cache xcache.Cache[string, string]
}

func (fs *CacheStore) Get(ctx context.Context, id string) Session {
	return &cacheSession{
		id:    id,
		cache: fs.Cache,
		ttl:   fs.TTL,
	}
}

var _ Session = (*cacheSession)(nil)

type cacheSession struct {
	id    string
	cache xcache.Cache[string, string]
	ttl   time.Duration
}

func (c *cacheSession) ID() string {
	return c.id
}

type cacheSessionMeta struct {
	CreatedAt int64    `json:"c"`
	Keys      []string `json:"k"`
}

func (c *cacheSession) metaKey() string {
	return c.id + ":m"
}

func (c *cacheSession) loadMeta(ctx context.Context) (*cacheSessionMeta, error) {
	k := c.metaKey()
	str, err := c.cache.Get(ctx, k)
	if err != nil && !xerror.IsNotFound(err) {
		return nil, err
	}
	var meta cacheSessionMeta
	err = json.Unmarshal([]byte(str), &meta)
	return &meta, err
}

func (c *cacheSession) updateMeta(ctx context.Context, addKeys []string, deleteKeys []string) error {
	k := c.metaKey()
	str, err := c.cache.Get(ctx, k)
	if err != nil && !xerror.IsNotFound(err) {
		return err
	}
	var meta cacheSessionMeta
	json.Unmarshal([]byte(str), &meta)
	now := time.Now().Unix()
	if meta.CreatedAt == 0 {
		meta.CreatedAt = now
	}
	meta.Keys = append(meta.Keys, addKeys...)
	meta.Keys = xslice.DeleteValue(meta.Keys, deleteKeys...)
	meta.Keys = xslice.Unique(meta.Keys)
	bf, _ := json.Marshal(meta)
	return c.cache.Set(ctx, k, string(bf), c.ttl)
}

func (c *cacheSession) dataKey(key string) string {
	return c.id + ":d:" + key
}

func (c *cacheSession) Set(ctx context.Context, key string, value string) error {
	err := c.updateMeta(ctx, []string{key}, nil)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, c.dataKey(key), value, c.ttl)
}

func (c *cacheSession) MSet(ctx context.Context, kv map[string]string) error {
	var errs []error
	for k, v := range kv {
		if err := c.Set(ctx, k, v); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c *cacheSession) Get(ctx context.Context, key string) (string, error) {
	val, err := c.cache.Get(ctx, c.dataKey(key))
	if err != nil && !xerror.IsNotFound(err) {
		return "", err
	}
	return val, nil
}

func (c *cacheSession) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	var errs []error
	for _, key := range keys {
		val, err := c.Get(ctx, key)
		if err != nil && !xerror.IsNotFound(err) {
			errs = append(errs, err)
		} else {
			result[key] = val
		}
	}
	return result, errors.Join(errs...)
}

func (c *cacheSession) Delete(ctx context.Context, keys ...string) error {
	newKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		newKeys = append(newKeys, c.dataKey(key))
	}
	err := c.cache.Delete(ctx, newKeys...)
	if err != nil {
		return err
	}
	return c.updateMeta(ctx, nil, keys)
}

func (c *cacheSession) Created(ctx context.Context) (time.Time, error) {
	meta, err := c.loadMeta(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(meta.CreatedAt, 0), nil
}

func (c *cacheSession) Save(ctx context.Context) error {
	return nil
}

func (c *cacheSession) Clear(ctx context.Context) error {
	meta, err := c.loadMeta(ctx)
	if err != nil {
		return err
	}
	if len(meta.Keys) == 0 {
		return nil
	}
	newKeys := make([]string, 0, len(meta.Keys))
	for _, key := range meta.Keys {
		newKeys = append(newKeys, c.dataKey(key))
	}
	err = c.cache.Delete(ctx, newKeys...)
	if err != nil {
		return err
	}
	return c.cache.Delete(ctx, c.metaKey())
}
