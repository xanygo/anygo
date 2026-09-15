package xkvmintor

import (
	"context"

	"github.com/xanygo/anygo/xkv"
)

var _ xkv.String[any] = (*monitorString[any])(nil)

type monitorString[V any] struct {
	key     string
	store   xkv.String[V]
	monitor *Monitor[V]
}

func (ms *monitorString[V]) Set(ctx context.Context, value V) error {
	err := ms.store.Set(ctx, value)
	ms.monitor.doAfter(ctx, DataTypeString, actionSet, err, ms.key)
	return err
}

func (ms *monitorString[V]) SetNX(ctx context.Context, value V) (bool, error) {
	ok, err := ms.store.SetNX(ctx, value)
	ms.monitor.doAfter(ctx, DataTypeString, actionSetNX, err, ms.key)
	return ok, err
}

func (ms *monitorString[V]) Get(ctx context.Context) (V, bool, error) {
	v, ok, err := ms.store.Get(ctx)
	ms.monitor.doAfter(ctx, DataTypeString, actionGet, err, ms.key)
	return v, ok, err
}

func (ms *monitorString[V]) GetDel(ctx context.Context) (V, bool, error) {
	v, ok, err := ms.store.GetDel(ctx)
	ms.monitor.doAfter(ctx, DataTypeString, actionGetDel, err, ms.key)
	return v, ok, err
}

func (ms *monitorString[V]) GetSet(ctx context.Context, value V) (V, bool, error) {
	v, ok, err := ms.store.GetSet(ctx, value)
	ms.monitor.doAfter(ctx, DataTypeString, actionGetSet, err, ms.key)
	return v, ok, err
}

func (ms *monitorString[V]) Incr(ctx context.Context) (int64, error) {
	v, err := ms.store.Incr(ctx)
	ms.monitor.doAfter(ctx, DataTypeString, actionIncr, err, ms.key)
	return v, err
}

func (ms *monitorString[V]) IncrBy(ctx context.Context, incr int64) (int64, error) {
	v, err := ms.store.IncrBy(ctx, incr)
	ms.monitor.doAfter(ctx, DataTypeString, actionIncrBy, err, ms.key)
	return v, err
}

func (ms *monitorString[V]) IncrByFloat(ctx context.Context, incr float64) (float64, error) {
	v, err := ms.store.IncrByFloat(ctx, incr)
	ms.monitor.doAfter(ctx, DataTypeString, actionIncrByFloat, err, ms.key)
	return v, err
}

func (ms *monitorString[V]) Decr(ctx context.Context) (int64, error) {
	v, err := ms.store.Decr(ctx)
	ms.monitor.doAfter(ctx, DataTypeString, actionDecr, err, ms.key)
	return v, err
}
