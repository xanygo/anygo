//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xkvx

import (
	"context"
	"errors"
	"io"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xredis"
)

var _ xkv.StringStorage = (*RedisStore)(nil)

// RedisStore 基于 redis 的 xkv StringStorage 实现
type RedisStore struct {
	KeyPrefix string // 可选，key 前缀
	Client    *xredis.Client
}

func (kv *RedisStore) String(key string) xkv.String[string] {
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
	return kvs.client.Set(ctx, kvs.key, value)
}

func (kvs *kvString) SetNX(ctx context.Context, value string) (bool, error) {
	return kvs.client.SetNX(ctx, kvs.key, value, 0)
}

func (kvs *kvString) Get(ctx context.Context) (string, bool, error) {
	value, err := kvs.client.Get(ctx, kvs.key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvs *kvString) GetSet(ctx context.Context, value string) (string, bool, error) {
	old, err := kvs.client.GetSet(ctx, kvs.key, value)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return old, err == nil, err
}

func (kvs *kvString) GetDel(ctx context.Context) (string, bool, error) {
	value, err := kvs.client.GetDel(ctx, kvs.key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvs *kvString) Incr(ctx context.Context) (int64, error) {
	return kvs.client.Incr(ctx, kvs.key)
}

func (kvs *kvString) IncrBy(ctx context.Context, incr int64) (int64, error) {
	return kvs.client.IncrBy(ctx, kvs.key, incr)
}

func (kvs *kvString) IncrByFloat(ctx context.Context, incr float64) (float64, error) {
	return kvs.client.IncrByFloat(ctx, kvs.key, incr)
}

func (kvs *kvString) Decr(ctx context.Context) (int64, error) {
	return kvs.client.Decr(ctx, kvs.key)
}

func (kv *RedisStore) List(key string) xkv.List[string] {
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

func (kv *RedisStore) Hash(key string) xkv.Hash[string] {
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

func (kvh *kvHash) HExists(ctx context.Context, field string) (bool, error) {
	return kvh.client.HExists(ctx, kvh.key, field)
}

func (kvh *kvHash) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	return kvh.client.HIncrBy(ctx, kvh.key, field, increment)
}

func (kvh *kvHash) HLen(ctx context.Context) (int64, error) {
	return kvh.client.HLen(ctx, kvh.key)
}

func (kv *RedisStore) Set(key string) xkv.Set[string] {
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

func (kvs *kvSet) SIsMember(ctx context.Context, member string) (bool, error) {
	return kvs.client.SIsMember(ctx, kvs.key, member)
}

func (kvs *kvSet) SMIsMember(ctx context.Context, members []string) ([]bool, error) {
	return kvs.client.SMIsMember(ctx, kvs.key, members...)
}

func (kvs *kvSet) SCard(ctx context.Context) (int64, error) {
	return kvs.client.SCard(ctx, kvs.key)
}

func (kvs *kvSet) SPop(ctx context.Context) (string, bool, error) {
	return kvs.client.SPop(ctx, kvs.key)
}

func (kvs *kvSet) SPopN(ctx context.Context, count int) ([]string, error) {
	return kvs.client.SPopN(ctx, kvs.key, count)
}

func (kvs *kvSet) SRandMember(ctx context.Context) (string, bool, error) {
	result, err := kvs.client.SRandMember(ctx, kvs.key, 1)
	if err != nil || len(result) != 1 {
		return "", false, err
	}
	return result[0], true, nil
}

func (kvs *kvSet) SRandMemberN(ctx context.Context, count int) ([]string, error) {
	return kvs.client.SRandMember(ctx, kvs.key, count)
}

func (kv *RedisStore) ZSet(key string) xkv.ZSet[string] {
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

func (kvz *kvZSet) ZIncrBy(ctx context.Context, score float64, member string) (float64, error) {
	num, err := kvz.client.ZIncrBy(ctx, kvz.key, score, member)
	return num, err
}

func (kvz *kvZSet) ZCount(ctx context.Context, min, max string) (int64, error) {
	num, err := kvz.client.ZCount(ctx, kvz.key, min, max)
	return num, err
}

func (kvz *kvZSet) ZLen(ctx context.Context) (int64, error) {
	num, err := kvz.client.ZCount(ctx, kvz.key, "-inf", "+inf")
	return num, err
}

func (kvz *kvZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	value, err := kvz.client.ZScore(ctx, kvz.key, member)
	if errors.Is(err, xredis.ErrNil) {
		return 0, false, nil
	}
	return value, err == nil, err
}

func (kvz *kvZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	return kvz.client.ZScanWalk(ctx, kvz.key, 0, "", 10, func(cursor uint64, items []xredis.Z) error {
		for _, item := range items {
			if !fn(item.Member, item.Score) {
				return io.EOF
			}
		}
		return nil
	})
}

func (kvz *kvZSet) ZRank(ctx context.Context, member string) (int64, float64, error) {
	index, score, err := kvz.client.ZRankWithScore(ctx, kvz.key, member)
	if errors.Is(err, xredis.ErrNil) {
		return -1, score, nil
	}
	return index, score, err
}

func (kvz *kvZSet) ZRem(ctx context.Context, members ...string) error {
	_, err := kvz.client.ZRem(ctx, kvz.key, members...)
	return err
}

func (kv *RedisStore) Delete(ctx context.Context, keys ...string) error {
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
