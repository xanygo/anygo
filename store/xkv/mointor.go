//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-15

package xkv

import (
	"context"
)

var _ Storage[any] = (*Monitor[any])(nil)

// Monitor 可以在 KV 操作完成后，执行监控回调
type Monitor[V any] struct {
	// 必填
	Store Storage[V]

	// 可选，操作完成后调用
	After func(ctx context.Context, typ string, action string, err error, keys ...string)
}

func (m *Monitor[V]) doAfter(ctx context.Context, typ string, action string, err error, keys ...string) {
	if m.After == nil {
		return
	}
	m.After(ctx, typ, action, err, keys...)
}

func (m *Monitor[V]) String(key string) String[V] {
	return &monitorString[V]{
		monitor: m,
		key:     key,
		store:   m.Store.String(key),
	}
}

var _ String[any] = (*monitorString[any])(nil)

type monitorString[V any] struct {
	key     string
	store   String[V]
	monitor *Monitor[V]
}

func (ms *monitorString[V]) Set(ctx context.Context, value V) error {
	err := ms.store.Set(ctx, value)
	ms.monitor.doAfter(ctx, "String", "Set", err, ms.key)
	return err
}

func (ms *monitorString[V]) SetNX(ctx context.Context, value V) (bool, error) {
	ok, err := ms.store.SetNX(ctx, value)
	ms.monitor.doAfter(ctx, "String", "SetNX", err, ms.key)
	return ok, err
}

func (ms *monitorString[V]) Get(ctx context.Context) (V, bool, error) {
	v, ok, err := ms.store.Get(ctx)
	ms.monitor.doAfter(ctx, "String", "Get", err, ms.key)
	return v, ok, err
}

func (ms *monitorString[V]) GetDel(ctx context.Context) (V, bool, error) {
	v, ok, err := ms.store.GetDel(ctx)
	ms.monitor.doAfter(ctx, "String", "GetDel", err, ms.key)
	return v, ok, err
}

func (ms *monitorString[V]) GetSet(ctx context.Context, value V) (V, bool, error) {
	v, ok, err := ms.store.GetSet(ctx, value)
	ms.monitor.doAfter(ctx, "String", "GetSet", err, ms.key)
	return v, ok, err
}

func (ms *monitorString[V]) Incr(ctx context.Context) (int64, error) {
	v, err := ms.store.Incr(ctx)
	ms.monitor.doAfter(ctx, "String", "Incr", err, ms.key)
	return v, err
}

func (ms *monitorString[V]) IncrBy(ctx context.Context, incr int64) (int64, error) {
	v, err := ms.store.IncrBy(ctx, incr)
	ms.monitor.doAfter(ctx, "String", "IncrBy", err, ms.key)
	return v, err
}

func (ms *monitorString[V]) IncrByFloat(ctx context.Context, incr float64) (float64, error) {
	v, err := ms.store.IncrByFloat(ctx, incr)
	ms.monitor.doAfter(ctx, "String", "IncrByFloat", err, ms.key)
	return v, err
}

func (ms *monitorString[V]) Decr(ctx context.Context) (int64, error) {
	v, err := ms.store.Decr(ctx)
	ms.monitor.doAfter(ctx, "String", "Decr", err, ms.key)
	return v, err
}

func (m *Monitor[V]) List(key string) List[V] {
	return &monitorList[V]{
		key:     key,
		store:   m.Store.List(key),
		monitor: m,
	}
}

var _ List[any] = (*monitorList[any])(nil)

type monitorList[V any] struct {
	key     string
	store   List[V]
	monitor *Monitor[V]
}

func (ml *monitorList[V]) LPush(ctx context.Context, values ...V) (int64, error) {
	val, err := ml.store.LPush(ctx, values...)
	ml.monitor.doAfter(ctx, "List", "LPush", err, ml.key)
	return val, err
}

func (ml *monitorList[V]) RPush(ctx context.Context, values ...V) (int64, error) {
	val, err := ml.store.RPush(ctx, values...)
	ml.monitor.doAfter(ctx, "List", "RPush", err, ml.key)
	return val, err
}

func (ml *monitorList[V]) LPop(ctx context.Context) (V, bool, error) {
	val, ok, err := ml.store.LPop(ctx)
	// 这个会修改数据，所以也是 write
	ml.monitor.doAfter(ctx, "List", "LPop", err, ml.key)
	return val, ok, err
}

func (ml *monitorList[V]) RPop(ctx context.Context) (V, bool, error) {
	val, ok, err := ml.store.RPop(ctx)
	// 这个会修改数据，所以也是 write
	ml.monitor.doAfter(ctx, "List", "RPop", err, ml.key)
	return val, ok, err
}

func (ml *monitorList[V]) LRem(ctx context.Context, count int64, element string) (int64, error) {
	val, err := ml.store.LRem(ctx, count, element)
	ml.monitor.doAfter(ctx, "List", "LRem", err, ml.key)
	return val, err
}

func (ml *monitorList[V]) Range(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.Range(ctx, fn)
	ml.monitor.doAfter(ctx, "List", "Range", err, ml.key)
	return err
}

func (ml *monitorList[V]) LRange(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.LRange(ctx, fn)
	ml.monitor.doAfter(ctx, "List", "LRange", err, ml.key)
	return err
}

func (ml *monitorList[V]) RRange(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.RRange(ctx, fn)
	ml.monitor.doAfter(ctx, "List", "RRange", err, ml.key)
	return err
}

func (ml *monitorList[V]) LLen(ctx context.Context) (int64, error) {
	num, err := ml.store.LLen(ctx)
	ml.monitor.doAfter(ctx, "List", "LLen", err, ml.key)
	return num, err
}

func (m *Monitor[V]) Hash(key string) Hash[V] {
	return &monitorHash[V]{
		key:     key,
		monitor: m,
		store:   m.Store.Hash(key),
	}
}

var _ Hash[any] = (*monitorHash[any])(nil)

type monitorHash[V any] struct {
	key     string
	store   Hash[V]
	monitor *Monitor[V]
}

func (mh *monitorHash[V]) HSet(ctx context.Context, field string, value V) error {
	err := mh.store.HSet(ctx, field, value)
	mh.monitor.doAfter(ctx, "Hash", "HSet", err, mh.key)
	return err
}

func (mh *monitorHash[V]) HMSet(ctx context.Context, data map[string]V) error {
	err := mh.store.HMSet(ctx, data)
	mh.monitor.doAfter(ctx, "Hash", "HMSet", err, mh.key)
	return err
}

func (mh *monitorHash[V]) HGet(ctx context.Context, field string) (V, bool, error) {
	val, ok, err := mh.store.HGet(ctx, field)
	mh.monitor.doAfter(ctx, "List", "HGet", err, mh.key)
	return val, ok, err
}

func (mh *monitorHash[V]) HDel(ctx context.Context, fields ...string) error {
	err := mh.store.HDel(ctx, fields...)
	mh.monitor.doAfter(ctx, "Hash", "HDel", err, mh.key)
	return err
}

func (mh *monitorHash[V]) HRange(ctx context.Context, fn func(field string, value V) bool) error {
	err := mh.store.HRange(ctx, fn)
	mh.monitor.doAfter(ctx, "List", "HRange", err, mh.key)
	return err
}

func (mh *monitorHash[V]) HGetAll(ctx context.Context) (map[string]V, error) {
	val, err := mh.store.HGetAll(ctx)
	mh.monitor.doAfter(ctx, "List", "HGetAll", err, mh.key)
	return val, err
}

func (mh *monitorHash[V]) HExists(ctx context.Context, field string) (bool, error) {
	found, err := mh.store.HExists(ctx, field)
	mh.monitor.doAfter(ctx, "List", "HExists", err, mh.key)
	return found, err
}

func (mh *monitorHash[V]) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	num, err := mh.store.HIncrBy(ctx, field, increment)
	mh.monitor.doAfter(ctx, "Hash", "HIncrBy", err, mh.key)
	return num, err
}

func (mh *monitorHash[V]) HLen(ctx context.Context) (int64, error) {
	num, err := mh.store.HLen(ctx)
	mh.monitor.doAfter(ctx, "List", "HLen", err, mh.key)
	return num, err
}

func (m *Monitor[V]) Set(key string) Set[V] {
	return &monitorSet[V]{
		monitor: m,
		key:     key,
		store:   m.Store.Set(key),
	}
}

var _ Set[any] = (*monitorSet[any])(nil)

type monitorSet[V any] struct {
	key     string
	store   Set[V]
	monitor *Monitor[V]
}

func (ms *monitorSet[V]) SAdd(ctx context.Context, members ...V) (int64, error) {
	val, err := ms.store.SAdd(ctx, members...)
	ms.monitor.doAfter(ctx, "Set", "SAdd", err, ms.key)
	return val, err
}

func (ms *monitorSet[V]) SRem(ctx context.Context, members ...V) error {
	err := ms.store.SRem(ctx, members...)
	ms.monitor.doAfter(ctx, "Set", "SRem", err, ms.key)
	return err
}

func (ms *monitorSet[V]) SRange(ctx context.Context, fn func(member V) bool) error {
	err := ms.store.SRange(ctx, fn)
	ms.monitor.doAfter(ctx, "Set", "SRange", err, ms.key)
	return err
}

func (ms *monitorSet[V]) SMembers(ctx context.Context) ([]V, error) {
	val, err := ms.store.SMembers(ctx)
	ms.monitor.doAfter(ctx, "Set", "SMembers", err, ms.key)
	return val, err
}

func (ms *monitorSet[V]) SCard(ctx context.Context) (int64, error) {
	val, err := ms.store.SCard(ctx)
	ms.monitor.doAfter(ctx, "Set", "SCard", err, ms.key)
	return val, err
}

func (ms *monitorSet[V]) SIsMember(ctx context.Context, member V) (bool, error) {
	found, err := ms.store.SIsMember(ctx, member)
	ms.monitor.doAfter(ctx, "Set", "SIsMember", err, ms.key)
	return found, err
}

func (ms *monitorSet[V]) SMIsMember(ctx context.Context, members []V) ([]bool, error) {
	result, err := ms.store.SMIsMember(ctx, members)
	ms.monitor.doAfter(ctx, "Set", "SMIsMember", err, ms.key)
	return result, err
}

func (ms *monitorSet[V]) SPop(ctx context.Context) (V, bool, error) {
	result, ok, err := ms.store.SPop(ctx)
	ms.monitor.doAfter(ctx, "Set", "SPop", err, ms.key)
	return result, ok, err
}

func (ms *monitorSet[V]) SPopN(ctx context.Context, count int) ([]V, error) {
	result, err := ms.store.SPopN(ctx, count)
	ms.monitor.doAfter(ctx, "Set", "SPopN", err, ms.key)
	return result, err
}

func (ms *monitorSet[V]) SRandMember(ctx context.Context) (V, bool, error) {
	result, ok, err := ms.store.SRandMember(ctx)
	ms.monitor.doAfter(ctx, "Set", "xSRandMemberx", err, ms.key)
	return result, ok, err
}

func (ms *monitorSet[V]) SRandMemberN(ctx context.Context, count int) ([]V, error) {
	result, err := ms.store.SRandMemberN(ctx, count)
	ms.monitor.doAfter(ctx, "Set", "SRandMemberN", err, ms.key)
	return result, err
}

func (m *Monitor[V]) ZSet(key string) ZSet[V] {
	return &monitorZSet[V]{
		store:   m.Store.ZSet(key),
		key:     key,
		monitor: m,
	}
}

var _ ZSet[any] = (*monitorZSet[any])(nil)

type monitorZSet[V any] struct {
	key     string
	store   ZSet[V]
	monitor *Monitor[V]
}

func (mz *monitorZSet[V]) ZAdd(ctx context.Context, score float64, member V) error {
	err := mz.store.ZAdd(ctx, score, member)
	mz.monitor.doAfter(ctx, "ZSet", "ZAdd", err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZScore(ctx context.Context, member V) (float64, bool, error) {
	val, ok, err := mz.store.ZScore(ctx, member)
	mz.monitor.doAfter(ctx, "ZSet", "ZScore", err, mz.key)
	return val, ok, err
}

func (mz *monitorZSet[V]) ZIncrBy(ctx context.Context, incr float64, member V) (float64, error) {
	val, err := mz.store.ZIncrBy(ctx, incr, member)
	mz.monitor.doAfter(ctx, "ZSet", "ZIncrBy", err, mz.key)
	return val, err
}

func (mz *monitorZSet[V]) ZRange(ctx context.Context, fn func(member V, score float64) bool) error {
	err := mz.store.ZRange(ctx, fn)
	mz.monitor.doAfter(ctx, "ZSet", "ZRange", err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZRangeByScore(ctx context.Context, min string, max string, fn func(member V, score float64) bool) error {
	err := mz.store.ZRangeByScore(ctx, min, max, fn)
	mz.monitor.doAfter(ctx, "ZSet", "ZRangeByScore", err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZRem(ctx context.Context, members ...V) error {
	err := mz.store.ZRem(ctx, members...)
	mz.monitor.doAfter(ctx, "ZSet", "ZRem", err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZRemRangeByScore(ctx context.Context, min, max string) (int64, error) {
	num, err := mz.store.ZRemRangeByScore(ctx, min, max)
	mz.monitor.doAfter(ctx, "ZSet", "ZRemRangeByScore", err, mz.key)
	return num, err
}

func (mz *monitorZSet[V]) ZCount(ctx context.Context, min, max string) (int64, error) {
	num, err := mz.store.ZCount(ctx, min, max)
	mz.monitor.doAfter(ctx, "ZSet", "ZCount", err, mz.key)
	return num, err
}

func (mz *monitorZSet[V]) ZLen(ctx context.Context) (int64, error) {
	num, err := mz.store.ZLen(ctx)
	mz.monitor.doAfter(ctx, "ZSet", "ZLen", err, mz.key)
	return num, err
}

func (mz *monitorZSet[V]) ZRank(ctx context.Context, member V) (int64, float64, error) {
	index, score, err := mz.store.ZRank(ctx, member)
	mz.monitor.doAfter(ctx, "ZSet", "ZRank", err, mz.key)
	return index, score, err
}

func (mz *monitorZSet[V]) ZPopMax(ctx context.Context, count int) ([]V, []float64, error) {
	values, scores, err := mz.store.ZPopMax(ctx, count)
	mz.monitor.doAfter(ctx, "ZSet", "ZPopMax", err, mz.key)
	return values, scores, err
}

func (mz *monitorZSet[V]) ZPopMin(ctx context.Context, count int) ([]V, []float64, error) {
	values, scores, err := mz.store.ZPopMin(ctx, count)
	mz.monitor.doAfter(ctx, "ZSet", "ZPopMin", err, mz.key)
	return values, scores, err
}

func (m *Monitor[V]) Delete(ctx context.Context, keys ...string) error {
	err := m.Store.Delete(ctx, keys...)
	m.doAfter(ctx, "Key", "Delete", err, keys...)
	return err
}

func (m *Monitor[V]) Has(ctx context.Context, key string) (bool, error) {
	val, err := m.Store.Has(ctx, key)
	m.doAfter(ctx, "Key", "Has", err, key)
	return val, err
}
