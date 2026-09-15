package xkvmintor

import (
	"context"

	"github.com/xanygo/anygo/xkv"
)

var _ xkv.Hash[any] = (*monitorHash[any])(nil)

type monitorHash[V any] struct {
	key     string
	store   xkv.Hash[V]
	monitor *Monitor[V]
}

func (mh *monitorHash[V]) HSet(ctx context.Context, field string, value V) error {
	err := mh.store.HSet(ctx, field, value)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHSet, err, mh.key)
	return err
}

func (mh *monitorHash[V]) HMSet(ctx context.Context, data map[string]V) error {
	err := mh.store.HMSet(ctx, data)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHMSet, err, mh.key)
	return err
}

func (mh *monitorHash[V]) HGet(ctx context.Context, field string) (V, bool, error) {
	val, ok, err := mh.store.HGet(ctx, field)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHGet, err, mh.key)
	return val, ok, err
}

func (mh *monitorHash[V]) HMGet(ctx context.Context, fields ...string) (map[string]V, error) {
	result, err := mh.store.HMGet(ctx, fields...)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHMGet, err, mh.key)
	return result, err
}

func (mh *monitorHash[V]) HDel(ctx context.Context, fields ...string) error {
	err := mh.store.HDel(ctx, fields...)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHDel, err, mh.key)
	return err
}

func (mh *monitorHash[V]) HRange(ctx context.Context, fn func(field string, value V) bool) error {
	err := mh.store.HRange(ctx, fn)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHRange, err, mh.key)
	return err
}

func (mh *monitorHash[V]) HGetAll(ctx context.Context) (map[string]V, error) {
	val, err := mh.store.HGetAll(ctx)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHGetAll, err, mh.key)
	return val, err
}

func (mh *monitorHash[V]) HExists(ctx context.Context, field string) (bool, error) {
	found, err := mh.store.HExists(ctx, field)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHExists, err, mh.key)
	return found, err
}

func (mh *monitorHash[V]) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	num, err := mh.store.HIncrBy(ctx, field, increment)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHIncrBy, err, mh.key)
	return num, err
}

func (mh *monitorHash[V]) HLen(ctx context.Context) (int64, error) {
	num, err := mh.store.HLen(ctx)
	mh.monitor.doAfter(ctx, DataTypeHash, actionHLen, err, mh.key)
	return num, err
}
