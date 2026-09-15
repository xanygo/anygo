package xkvmintor

import (
	"context"

	"github.com/xanygo/anygo/xkv"
)

var _ xkv.Set[any] = (*monitorSet[any])(nil)

type monitorSet[V any] struct {
	key     string
	store   xkv.Set[V]
	monitor *Monitor[V]
}

func (ms *monitorSet[V]) SAdd(ctx context.Context, members ...V) (int64, error) {
	val, err := ms.store.SAdd(ctx, members...)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSAdd, err, ms.key)
	return val, err
}

func (ms *monitorSet[V]) SRem(ctx context.Context, members ...V) error {
	err := ms.store.SRem(ctx, members...)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSRem, err, ms.key)
	return err
}

func (ms *monitorSet[V]) SRange(ctx context.Context, fn func(member V) bool) error {
	err := ms.store.SRange(ctx, fn)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSRange, err, ms.key)
	return err
}

func (ms *monitorSet[V]) SMembers(ctx context.Context) ([]V, error) {
	val, err := ms.store.SMembers(ctx)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSMembers, err, ms.key)
	return val, err
}

func (ms *monitorSet[V]) SCard(ctx context.Context) (int64, error) {
	val, err := ms.store.SCard(ctx)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSCard, err, ms.key)
	return val, err
}

func (ms *monitorSet[V]) SIsMember(ctx context.Context, member V) (bool, error) {
	found, err := ms.store.SIsMember(ctx, member)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSIsMember, err, ms.key)
	return found, err
}

func (ms *monitorSet[V]) SMIsMember(ctx context.Context, members []V) ([]bool, error) {
	result, err := ms.store.SMIsMember(ctx, members)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSMIsMember, err, ms.key)
	return result, err
}

func (ms *monitorSet[V]) SPop(ctx context.Context) (V, bool, error) {
	result, ok, err := ms.store.SPop(ctx)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSPop, err, ms.key)
	return result, ok, err
}

func (ms *monitorSet[V]) SPopN(ctx context.Context, count int) ([]V, error) {
	result, err := ms.store.SPopN(ctx, count)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSPopN, err, ms.key)
	return result, err
}

func (ms *monitorSet[V]) SRandMember(ctx context.Context) (V, bool, error) {
	result, ok, err := ms.store.SRandMember(ctx)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSRandMember, err, ms.key)
	return result, ok, err
}

func (ms *monitorSet[V]) SRandMemberN(ctx context.Context, count int) ([]V, error) {
	result, err := ms.store.SRandMemberN(ctx, count)
	ms.monitor.doAfter(ctx, DataTypeSet, actionSRandMemberN, err, ms.key)
	return result, err
}
