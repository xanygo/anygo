//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xkvx

import (
	"context"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/internal/rds"
	"github.com/xanygo/anygo/store/xredis"
)

var _ xkv.StringStorage = (*RedisStore)(nil)

// RedisStore 基于 redis 的 xkv StringStorage 实现
type RedisStore struct {
	KeyPrefix string // 可选，key 前缀
	Client    *xredis.Client
}

func (kv *RedisStore) String(key string) xkv.String[string] {
	return &rds.String{
		Key:    kv.KeyPrefix + key,
		Client: kv.Client,
	}
}

func (kv *RedisStore) List(key string) xkv.List[string] {
	return &rds.List{
		Key:    kv.KeyPrefix + key,
		Client: kv.Client,
	}
}

func (kv *RedisStore) Hash(key string) xkv.Hash[string] {
	return &rds.Hash{
		Client: kv.Client,
		Key:    kv.KeyPrefix + key,
	}
}

func (kv *RedisStore) Set(key string) xkv.Set[string] {
	return &rds.Set{
		Client: kv.Client,
		Key:    kv.KeyPrefix + key,
	}
}

func (kv *RedisStore) ZSet(key string) xkv.ZSet[string] {
	return &rds.ZSet{
		Client: kv.Client,
		Key:    kv.KeyPrefix + key,
	}
}

func (kv *RedisStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if kv.KeyPrefix != "" {
		for i := range keys {
			keys[i] = kv.KeyPrefix + keys[i]
		}
	}
	_, err := kv.Client.Del(ctx, keys...)
	return err
}

func (kv *RedisStore) Has(ctx context.Context, key string) (bool, error) {
	num, err := kv.Client.EXISTS(ctx, kv.KeyPrefix+key)
	return num == 1, err
}
