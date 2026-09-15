package xkvmintor

import (
	"context"

	"github.com/xanygo/anygo/xkv"
)

var _ xkv.ZSet[any] = (*monitorZSet[any])(nil)

type monitorZSet[V any] struct {
	key     string
	store   xkv.ZSet[V]
	monitor *Monitor[V]
}

func (mz *monitorZSet[V]) ZAdd(ctx context.Context, score float64, member V) error {
	err := mz.store.ZAdd(ctx, score, member)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZAdd, err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZMAdd(ctx context.Context, items ...xkv.ZItem[V]) error {
	err := mz.store.ZMAdd(ctx, items...)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZAdd, err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZScore(ctx context.Context, member V) (float64, bool, error) {
	val, ok, err := mz.store.ZScore(ctx, member)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZScore, err, mz.key)
	return val, ok, err
}

func (mz *monitorZSet[V]) ZIncrBy(ctx context.Context, incr float64, member V) (float64, error) {
	val, err := mz.store.ZIncrBy(ctx, incr, member)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZIncrBy, err, mz.key)
	return val, err
}

func (mz *monitorZSet[V]) ZRange(ctx context.Context, fn func(member V, score float64) bool) error {
	err := mz.store.ZRange(ctx, fn)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZRange, err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZRangeByScore(ctx context.Context, min string, max string, fn func(member V, score float64) bool) error {
	err := mz.store.ZRangeByScore(ctx, min, max, fn)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZRangeByScore, err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZRem(ctx context.Context, members ...V) error {
	err := mz.store.ZRem(ctx, members...)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZRem, err, mz.key)
	return err
}

func (mz *monitorZSet[V]) ZRemRangeByScore(ctx context.Context, min, max string) (int64, error) {
	num, err := mz.store.ZRemRangeByScore(ctx, min, max)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZRemRangeByScore, err, mz.key)
	return num, err
}

func (mz *monitorZSet[V]) ZCount(ctx context.Context, min, max string) (int64, error) {
	num, err := mz.store.ZCount(ctx, min, max)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZCount, err, mz.key)
	return num, err
}

func (mz *monitorZSet[V]) ZLen(ctx context.Context) (int64, error) {
	num, err := mz.store.ZLen(ctx)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZLen, err, mz.key)
	return num, err
}

func (mz *monitorZSet[V]) ZRank(ctx context.Context, member V) (int64, float64, error) {
	index, score, err := mz.store.ZRank(ctx, member)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZRank, err, mz.key)
	return index, score, err
}

func (mz *monitorZSet[V]) ZPopMax(ctx context.Context, count int) ([]V, []float64, error) {
	values, scores, err := mz.store.ZPopMax(ctx, count)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZPopMax, err, mz.key)
	return values, scores, err
}

func (mz *monitorZSet[V]) ZPopMin(ctx context.Context, count int) ([]V, []float64, error) {
	values, scores, err := mz.store.ZPopMin(ctx, count)
	mz.monitor.doAfter(ctx, DataTypeZSet, actionZPopMin, err, mz.key)
	return values, scores, err
}

func (m *Monitor[V]) Delete(ctx context.Context, keys ...string) error {
	err := m.Store.Delete(ctx, keys...)
	m.doAfter(ctx, DataTypeKey, actionDelete, err, keys...)
	return err
}

func (m *Monitor[V]) Has(ctx context.Context, key string) (bool, error) {
	val, err := m.Store.Has(ctx, key)
	m.doAfter(ctx, DataTypeKey, actionHas, err, key)
	return val, err
}
