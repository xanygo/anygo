//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpool

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zdefine"
	"github.com/xanygo/anygo/xpp"
)

type GroupKey[T comparable] interface {
	Key() T
}

var _ GroupKey[string] = (*groupKey)(nil)
var _ zdefine.HasKey[string] = (*groupKey)(nil)

// 此类型只用于做类型约定检查
type groupKey struct {
	key string
}

func (g groupKey) Key() string {
	return g.key
}

func NewGroupPool[K GroupKey[T], T comparable, V io.Closer](opt *Option, fac GroupFactory[K, T, V]) GroupPool[K, T, V] {
	opt = opt.Normalization()
	bgCtx, bgCancel := context.WithCancel(context.Background())
	return &simpleGroup[K, T, V]{
		option:  opt,
		creator: fac,
		bgCtx:   bgCtx,
		stop:    bgCancel,
	}
}

type GroupPool[K GroupKey[T], T comparable, V io.Closer] interface {
	Get(ctx context.Context, key K) (Entry[V], error)
	Put(e Entry[V], err error)
	Close() error

	Stats() Stats
	GroupStats() map[T]Stats
}

type GroupFactory[K GroupKey[T], T comparable, V io.Closer] interface {
	NewWithKey(ctx context.Context, key K) (V, error)
}

type GroupFactoryFunc[K comparable, V io.Closer] func(ctx context.Context, key K) (V, error)

func (gcf GroupFactoryFunc[K, V]) NewWithKey(ctx context.Context, key K) (V, error) {
	return gcf(ctx, key)
}

var _ GroupPool[groupKey, string, net.Conn] = (*simpleGroup[groupKey, string, net.Conn])(nil)

type simpleGroup[K GroupKey[T], T comparable, V io.Closer] struct {
	option   *Option
	pools    map[T]Pool[V]
	lastUsed map[T]time.Time
	mux      sync.Mutex
	creator  GroupFactory[K, T, V]
	solo     xpp.SoloTask

	bgCtx context.Context
	stop  context.CancelFunc
}

func (group *simpleGroup[K, T, V]) getPool(key K) Pool[V] {
	group.mux.Lock()
	defer group.mux.Unlock()
	if group.pools == nil {
		group.pools = make(map[T]Pool[V], 4)
	}
	if group.lastUsed == nil {
		group.lastUsed = make(map[T]time.Time, 4)
	}
	ks := key.Key()
	group.lastUsed[ks] = time.Now()
	p, ok := group.pools[ks]
	if ok {
		return p
	}
	p = group.newPool(key)
	group.pools[ks] = p
	return p
}

func (group *simpleGroup[K, T, V]) newPool(key K) Pool[V] {
	ct := FactoryFunc[V](func(ctx context.Context) (V, error) {
		return group.creator.NewWithKey(ctx, key)
	})
	return New[V](group.option, ct)
}

func (group *simpleGroup[K, T, V]) Get(ctx context.Context, key K) (Entry[V], error) {
	group.solo.RunContext(group.bgCtx, group.clearEmpty, 5*time.Minute, 10*time.Second)
	p := group.getPool(key)
	return p.Get(ctx)
}

func (group *simpleGroup[K, T, V]) Put(e Entry[V], err error) {
	group.solo.RunContext(group.bgCtx, group.clearEmpty, 5*time.Minute, 10*time.Second)
	e.Release(err)
}

func (group *simpleGroup[K, T, V]) Close() error {
	group.mux.Lock()
	defer group.mux.Unlock()
	var errs []error
	for _, pool := range group.pools {
		if err := pool.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	group.stop()
	return errors.Join(errs...)
}

func (group *simpleGroup[K, T, V]) clearEmpty(ctx context.Context) {
	group.mux.Lock()
	names := xmap.Keys(group.pools)
	group.mux.Unlock()
	for _, name := range names {
		group.checkAndClearOne(name)
	}
}

func (group *simpleGroup[K, T, V]) checkAndClearOne(name T) {
	group.mux.Lock()

	pool, ok1 := group.pools[name]
	if !ok1 {
		group.mux.Unlock()
		return
	}
	tm := group.lastUsed[name]
	if time.Now().Before(tm.Add(group.option.MaxPoolIdleTime)) {
		group.mux.Unlock()
		return
	}
	st := pool.Stats()
	canClose := st.InUse == 0 && st.NumOpen == 0
	if !canClose {
		group.mux.Unlock()
		return
	}
	delete(group.pools, name)
	delete(group.lastUsed, name)
	group.mux.Unlock()

	_ = pool.Close()
}

func (group *simpleGroup[K, T, V]) Stats() Stats {
	group.mux.Lock()
	defer group.mux.Unlock()
	var st Stats
	for _, pool := range group.pools {
		st = st.Add(pool.Stats())
	}
	return st
}

func (group *simpleGroup[K, T, V]) GroupStats() map[T]Stats {
	group.mux.Lock()
	defer group.mux.Unlock()
	result := make(map[T]Stats, len(group.pools))
	for key, pool := range group.pools {
		result[key] = pool.Stats()
	}
	return result
}
