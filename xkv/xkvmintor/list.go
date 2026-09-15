package xkvmintor

import (
	"context"

	"github.com/xanygo/anygo/xkv"
)

var _ xkv.List[any] = (*monitorList[any])(nil)

type monitorList[V any] struct {
	key     string
	store   xkv.List[V]
	monitor *Monitor[V]
}

func (ml *monitorList[V]) LPush(ctx context.Context, values ...V) (int64, error) {
	val, err := ml.store.LPush(ctx, values...)
	ml.monitor.doAfter(ctx, DataTypeList, actionLPush, err, ml.key)
	return val, err
}

func (ml *monitorList[V]) RPush(ctx context.Context, values ...V) (int64, error) {
	val, err := ml.store.RPush(ctx, values...)
	ml.monitor.doAfter(ctx, DataTypeList, actionRPush, err, ml.key)
	return val, err
}

func (ml *monitorList[V]) LPop(ctx context.Context) (V, bool, error) {
	val, ok, err := ml.store.LPop(ctx)
	ml.monitor.doAfter(ctx, DataTypeList, actionLPop, err, ml.key)
	return val, ok, err
}

func (ml *monitorList[V]) LPopN(ctx context.Context, count int) ([]V, error) {
	items, err := ml.store.LPopN(ctx, count)
	ml.monitor.doAfter(ctx, DataTypeList, actionLPopN, err, ml.key)
	return items, err
}

func (ml *monitorList[V]) RPop(ctx context.Context) (V, bool, error) {
	val, ok, err := ml.store.RPop(ctx)
	// 这个会修改数据，所以也是 write
	ml.monitor.doAfter(ctx, DataTypeList, actionRPop, err, ml.key)
	return val, ok, err
}

func (ml *monitorList[V]) RPopN(ctx context.Context, count int) ([]V, error) {
	items, err := ml.store.RPopN(ctx, count)
	ml.monitor.doAfter(ctx, DataTypeList, actionRPopN, err, ml.key)
	return items, err
}

func (ml *monitorList[V]) LRem(ctx context.Context, count int64, element string) (int64, error) {
	val, err := ml.store.LRem(ctx, count, element)
	ml.monitor.doAfter(ctx, DataTypeList, actionLRem, err, ml.key)
	return val, err
}

func (ml *monitorList[V]) Range(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.Range(ctx, fn)
	ml.monitor.doAfter(ctx, DataTypeList, actionRange, err, ml.key)
	return err
}

func (ml *monitorList[V]) LRange(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.LRange(ctx, fn)
	ml.monitor.doAfter(ctx, DataTypeList, actionLRange, err, ml.key)
	return err
}

func (ml *monitorList[V]) RRange(ctx context.Context, fn func(val V) bool) error {
	err := ml.store.RRange(ctx, fn)
	ml.monitor.doAfter(ctx, DataTypeList, actionRRange, err, ml.key)
	return err
}

func (ml *monitorList[V]) LLen(ctx context.Context) (int64, error) {
	num, err := ml.store.LLen(ctx)
	ml.monitor.doAfter(ctx, DataTypeList, actionLLen, err, ml.key)
	return num, err
}
