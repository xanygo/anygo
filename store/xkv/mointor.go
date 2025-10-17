//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-15

package xkv

import "context"

var _ Storage[any] = (*Monitor[any])(nil)

// Monitor 可以在 KV 操作完成后，执行监控回调
type Monitor[V any] struct {
	Store       Storage[V]                                       // 必填
	AfterRead   func(ctx context.Context, key string, err error) // 可选，读操作完成后调用
	AfterWrite  func(ctx context.Context, key string, err error) // 可选，写操作完成后调用
	AfterDelete func(ctx context.Context, key string, err error) // 可选，key删除后调用
}

func (m *Monitor[V]) doAfterRead(ctx context.Context, key string, err error) {
	if m.AfterRead == nil {
		return
	}
	m.AfterRead(ctx, key, err)
}

func (m *Monitor[V]) doAfterWrite(ctx context.Context, key string, err error) {
	if m.AfterWrite == nil {
		return
	}
	m.AfterWrite(ctx, key, err)
}

func (m *Monitor[V]) doAfterDelete(ctx context.Context, key string, err error) {
	if m.AfterDelete == nil {
		return
	}
	m.AfterDelete(ctx, key, err)
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
	ms.monitor.doAfterWrite(ctx, ms.key, err)
	return err
}

func (ms *monitorString[V]) Get(ctx context.Context) (V, bool, error) {
	v, ok, err := ms.store.Get(ctx)
	ms.monitor.doAfterRead(ctx, ms.key, err)
	return v, ok, err
}

func (ms *monitorString[V]) Incr(ctx context.Context) (int64, error) {
	v, err := ms.store.Incr(ctx)
	ms.monitor.doAfterWrite(ctx, ms.key, err)
	return v, err
}

func (ms *monitorString[V]) Decr(ctx context.Context) (int64, error) {
	v, err := ms.store.Decr(ctx)
	ms.monitor.doAfterWrite(ctx, ms.key, err)
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
	ml.monitor.doAfterWrite(ctx, ml.key, err)
	return val, err
}

func (ml *monitorList[V]) RPush(ctx context.Context, values ...V) (int64, error) {
	val, err := ml.store.RPush(ctx, values...)
	ml.monitor.doAfterWrite(ctx, ml.key, err)
	return val, err
}

func (ml *monitorList[V]) LPop(ctx context.Context) (V, bool, error) {
	val, ok, err := ml.store.LPop(ctx)
	// 这个会修改数据，所以也是 write
	ml.monitor.doAfterWrite(ctx, ml.key, err)
	return val, ok, err
}

func (ml *monitorList[V]) RPop(ctx context.Context) (V, bool, error) {
	val, ok, err := ml.store.RPop(ctx)
	// 这个会修改数据，所以也是 write
	ml.monitor.doAfterWrite(ctx, ml.key, err)
	return val, ok, err
}

func (ml *monitorList[V]) LRem(ctx context.Context, count int64, element string) (int64, error) {
	val, err := ml.store.LRem(ctx, count, element)
	ml.monitor.doAfterWrite(ctx, ml.key, err)
	return val, err
}

func (ml *monitorList[V]) Range(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.Range(ctx, fn)
	ml.monitor.doAfterRead(ctx, ml.key, err)
	return err
}

func (ml *monitorList[V]) LRange(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.LRange(ctx, fn)
	ml.monitor.doAfterRead(ctx, ml.key, err)
	return err
}

func (ml *monitorList[V]) RRange(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.RRange(ctx, fn)
	ml.monitor.doAfterRead(ctx, ml.key, err)
	return err
}

func (ml *monitorList[V]) LLen(ctx context.Context) (int64, error) {
	num, err := ml.store.LLen(ctx)
	ml.monitor.doAfterRead(ctx, ml.key, err)
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
	mh.monitor.doAfterWrite(ctx, mh.key, err)
	return err
}

func (mh *monitorHash[V]) HMSet(ctx context.Context, data map[string]V) error {
	err := mh.store.HMSet(ctx, data)
	mh.monitor.doAfterWrite(ctx, mh.key, err)
	return err
}

func (mh *monitorHash[V]) HGet(ctx context.Context, field string) (V, bool, error) {
	val, ok, err := mh.store.HGet(ctx, field)
	mh.monitor.doAfterRead(ctx, mh.key, err)
	return val, ok, err
}

func (mh *monitorHash[V]) HDel(ctx context.Context, fields ...string) error {
	err := mh.store.HDel(ctx, fields...)
	mh.monitor.doAfterWrite(ctx, mh.key, err)
	return err
}

func (mh *monitorHash[V]) HRange(ctx context.Context, fn func(field string, value V) bool) error {
	err := mh.store.HRange(ctx, fn)
	mh.monitor.doAfterRead(ctx, mh.key, err)
	return err
}

func (mh *monitorHash[V]) HGetAll(ctx context.Context) (map[string]V, error) {
	val, err := mh.store.HGetAll(ctx)
	mh.monitor.doAfterRead(ctx, mh.key, err)
	return val, err
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
	ms.monitor.doAfterWrite(ctx, ms.key, err)
	return val, err
}

func (ms *monitorSet[V]) SRem(ctx context.Context, members ...V) error {
	err := ms.store.SRem(ctx, members...)
	ms.monitor.doAfterWrite(ctx, ms.key, err)
	return err
}

func (ms *monitorSet[V]) SRange(ctx context.Context, fn func(member V) bool) error {
	err := ms.store.SRange(ctx, fn)
	ms.monitor.doAfterRead(ctx, ms.key, err)
	return err
}

func (ms *monitorSet[V]) SMembers(ctx context.Context) ([]V, error) {
	val, err := ms.store.SMembers(ctx)
	ms.monitor.doAfterRead(ctx, ms.key, err)
	return val, err
}

func (ms *monitorSet[V]) SCard(ctx context.Context) (int64, error) {
	val, err := ms.store.SCard(ctx)
	ms.monitor.doAfterRead(ctx, ms.key, err)
	return val, err
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
	mz.monitor.doAfterWrite(ctx, mz.key, err)
	return err
}

func (mz *monitorZSet[V]) ZScore(ctx context.Context, member V) (float64, bool, error) {
	val, ok, err := mz.store.ZScore(ctx, member)
	mz.monitor.doAfterRead(ctx, mz.key, err)
	return val, ok, err
}

func (mz *monitorZSet[V]) ZRange(ctx context.Context, fn func(member V, score float64) bool) error {
	err := mz.store.ZRange(ctx, fn)
	mz.monitor.doAfterRead(ctx, mz.key, err)
	return err
}

func (mz *monitorZSet[V]) ZRem(ctx context.Context, members ...V) error {
	err := mz.store.ZRem(ctx, members...)
	mz.monitor.doAfterWrite(ctx, mz.key, err)
	return err
}

func (m *Monitor[V]) Delete(ctx context.Context, keys ...string) error {
	err := m.Store.Delete(ctx, keys...)
	for _, key := range keys {
		m.doAfterDelete(ctx, key, err)
	}
	return err
}
