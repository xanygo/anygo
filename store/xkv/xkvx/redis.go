//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xkvx

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/internal/rds"
	"github.com/xanygo/anygo/store/xredis"
)

var _ xkv.StringStorage = (*Redis)(nil)

// Redis 基于 redis 的 xkv StringStorage 实现
type Redis struct {
	KeyPrefix string // 可选，key 前缀
	Client    *xredis.Client
}

func (kv *Redis) Init(param map[string]any) error {
	if kv.KeyPrefix == "" {
		kv.KeyPrefix, _ = xmap.GetString(param, "KeyPrefix")
	}
	if kv.Client == nil {
		name, _ := xmap.GetString(param, "Service")
		if name == "" {
			return fmt.Errorf("invalid Service in %v", param)
		}
		kv.Client = xredis.NewClient(name)
	}
	return nil
}

func (kv *Redis) String(key string) xkv.String[string] {
	return &rds.String{
		Key:    kv.KeyPrefix + key,
		Client: kv.Client,
	}
}

func (kv *Redis) List(key string) xkv.List[string] {
	return &rds.List{
		Key:    kv.KeyPrefix + key,
		Client: kv.Client,
	}
}

func (kv *Redis) Hash(key string) xkv.Hash[string] {
	return &rds.Hash{
		Client: kv.Client,
		Key:    kv.KeyPrefix + key,
	}
}

func (kv *Redis) Set(key string) xkv.Set[string] {
	return &rds.Set{
		Client: kv.Client,
		Key:    kv.KeyPrefix + key,
	}
}

func (kv *Redis) ZSet(key string) xkv.ZSet[string] {
	return &rds.ZSet{
		Client: kv.Client,
		Key:    kv.KeyPrefix + key,
	}
}

func (kv *Redis) Delete(ctx context.Context, keys ...string) error {
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

func (kv *Redis) Has(ctx context.Context, key string) (bool, error) {
	num, err := kv.Client.EXISTS(ctx, kv.KeyPrefix+key)
	return num == 1, err
}
