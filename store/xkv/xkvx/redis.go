//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xkvx

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xredis"
)

var _ xkv.StringStorage = (*RedisStorage)(nil)

// RedisStorage 基于 redis 的 xkv StringStorage 实现
type RedisStorage struct {
	KeyPrefix string
	Client    *xredis.Client
}

func (kv *RedisStorage) String(key string) xkv.String[string] {
	return &kvString{
		key:    kv.KeyPrefix + key,
		client: kv.Client,
	}
}

var _ xkv.String[string] = (*kvString)(nil)

type kvString struct {
	client *xredis.Client
	key    string
}

func (kvs *kvString) Set(ctx context.Context, value string) error {
	return kvs.client.Set(ctx, kvs.key, value, 0)
}

func (kvs *kvString) Get(ctx context.Context) (string, bool, error) {
	value, err := kvs.client.Get(ctx, kvs.key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvs *kvString) Incr(ctx context.Context) (int64, error) {
	return kvs.client.Incr(ctx, kvs.key)
}

func (kvs *kvString) Decr(ctx context.Context) (int64, error) {
	return kvs.client.Decr(ctx, kvs.key)
}

func (kv *RedisStorage) List(key string) xkv.List[string] {
	return &kvList{
		key:    kv.KeyPrefix + key,
		client: kv.Client,
	}
}

var _ xkv.List[string] = (*kvList)(nil)

type kvList struct {
	client *xredis.Client
	key    string
}

func (kvl *kvList) LPush(ctx context.Context, values ...string) (int64, error) {
	return kvl.client.LPush(ctx, kvl.key, values...)
}

func (kvl *kvList) RPush(ctx context.Context, values ...string) (int64, error) {
	return kvl.client.RPush(ctx, kvl.key, values...)
}

func (kvl *kvList) LPop(ctx context.Context) (string, bool, error) {
	value, err := kvl.client.LPop(ctx, kvl.key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvl *kvList) RPop(ctx context.Context) (string, bool, error) {
	value, err := kvl.client.RPop(ctx, kvl.key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvl *kvList) LRem(ctx context.Context, count int64, element string) (int64, error) {
	return kvl.client.LRem(ctx, kvl.key, count, element)
}

func (kvl *kvList) Range(ctx context.Context, fn func(val string) bool) error {
	return kvl.LRange(ctx, fn)
}

func (kvl *kvList) LRange(ctx context.Context, fn func(val string) bool) error {
	for start := int64(0); ; start += 10 {
		stop := start + 10
		values, err := kvl.client.LRange(ctx, kvl.key, start, stop)
		if errors.Is(err, xredis.ErrNil) || len(values) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		for _, val := range values {
			if !fn(val) {
				return nil
			}
		}
	}
}

func (kvl *kvList) RRange(ctx context.Context, fn func(val string) bool) error {
	for stop := int64(-1); ; stop -= 9 {
		start := stop - 9
		values, err := kvl.client.LRange(ctx, kvl.key, start, stop)
		if errors.Is(err, xredis.ErrNil) || len(values) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		for i := len(values) - 1; i >= 0; i-- {
			if !fn(values[i]) {
				return nil
			}
		}
	}
}

func (kvl *kvList) LLen(ctx context.Context) (int64, error) {
	return kvl.client.LLen(ctx, kvl.key)
}

func (kv *RedisStorage) Hash(key string) xkv.Hash[string] {
	return &kvHash{
		client: kv.Client,
		key:    kv.KeyPrefix + key,
	}
}

var _ xkv.Hash[string] = (*kvHash)(nil)

type kvHash struct {
	client *xredis.Client
	key    string
}

func (kvh *kvHash) HSet(ctx context.Context, field string, value string) error {
	_, err := kvh.client.HSet(ctx, kvh.key, field, value)
	return err
}

func (kvh *kvHash) HMSet(ctx context.Context, values map[string]string) error {
	return kvh.client.HMSet(ctx, kvh.key, values)
}

func (kvh *kvHash) HGet(ctx context.Context, field string) (string, bool, error) {
	value, err := kvh.client.HGet(ctx, kvh.key, field)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvh *kvHash) HDel(ctx context.Context, fields ...string) error {
	_, err := kvh.client.HDel(ctx, kvh.key, fields...)
	return err
}

func (kvh *kvHash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	// todo: scan
	values, err := kvh.HGetAll(ctx)
	if err != nil {
		return err
	}
	for k, v := range values {
		if !fn(k, v) {
			return nil
		}
	}
	return nil
}

func (kvh *kvHash) HGetAll(ctx context.Context) (map[string]string, error) {
	return kvh.client.HGetAll(ctx, kvh.key)
}

func (kv *RedisStorage) Set(key string) xkv.Set[string] {
	return &kvSet{
		client: kv.Client,
		key:    kv.KeyPrefix + key,
	}
}

var _ xkv.Set[string] = (*kvSet)(nil)

type kvSet struct {
	client *xredis.Client
	key    string
}

func (kvs *kvSet) SAdd(ctx context.Context, members ...string) (int64, error) {
	return kvs.client.SAdd(ctx, kvs.key, members...)
}

func (kvs *kvSet) SRem(ctx context.Context, members ...string) error {
	_, err := kvs.client.SRem(ctx, kvs.key, members...)
	return err
}

func (kvs *kvSet) SRange(ctx context.Context, fn func(val string) bool) error {
	// todo: sscan
	values, err := kvs.SMembers(ctx)
	if err != nil {
		return nil
	}
	for _, val := range values {
		if !fn(val) {
			return nil
		}
	}
	return nil
}

func (kvs *kvSet) SMembers(ctx context.Context) ([]string, error) {
	values, err := kvs.client.SMembers(ctx, kvs.key)
	if errors.Is(err, xredis.ErrNil) {
		return nil, nil
	}
	return values, err
}

func (kvs *kvSet) SCard(ctx context.Context) (int64, error) {
	return kvs.client.SCard(ctx, kvs.key)
}

func (kv *RedisStorage) ZSet(key string) xkv.ZSet[string] {
	return &kvZSet{
		client: kv.Client,
		key:    kv.KeyPrefix + key,
	}
}

var _ xkv.ZSet[string] = (*kvZSet)(nil)

type kvZSet struct {
	client *xredis.Client
	key    string
}

func (kvz *kvZSet) ZAdd(ctx context.Context, score float64, member string) error {
	_, err := kvz.client.ZAdd(ctx, kvz.key, score, member)
	return err
}

func (kvz *kvZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	value, err := kvz.client.ZScore(ctx, kvz.key, member)
	if errors.Is(err, xredis.ErrNil) {
		return 0, false, nil
	}
	return value, err == nil, err
}

func (kvz *kvZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	for start := int64(0); ; start += 10 {
		stop := start + 9
		values, err := kvz.client.ZRangeWithScore(ctx, kvz.key, start, stop)
		if errors.Is(err, xredis.ErrNil) || len(values) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		for _, item := range values {
			if !fn(item.Member, item.Score) {
				return nil
			}
		}
	}
}

func (kvz *kvZSet) ZRem(ctx context.Context, members ...string) error {
	_, err := kvz.client.ZRem(ctx, kvz.key, members...)
	return err
}

func (kv *RedisStorage) Delete(ctx context.Context, keys ...string) error {
	if kv.KeyPrefix != "" {
		for i := 0; i < len(keys); i++ {
			keys[i] = kv.KeyPrefix + keys[i]
		}
	}
	_, err := kv.Client.Del(ctx, keys...)
	return err
}
