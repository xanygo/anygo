//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xpool

import (
	"context"
	"io"
	"sync/atomic"
)

var _ Pool[io.Closer] = (*DirectPool[io.Closer])(nil)

// DirectPool 实现 Pool 接口，但是每次调用都会返回一个新元素
type DirectPool[V io.Closer] struct {
	Factory   Factory[V]
	numOpened atomic.Int64
}

func (d *DirectPool[V]) Get(ctx context.Context) (Entry[V], error) {
	item, err := d.Factory.New(ctx)
	if err != nil {
		return nil, err
	}
	et := NewOpenEntry(item, d)
	et.UpdateUsing()
	d.numOpened.Add(1)
	return et, nil
}

func (d *DirectPool[V]) Put(e Entry[V], err error) {
	e.Close()
	d.numOpened.Add(-1)
}

func (d *DirectPool[V]) Close() error {
	return nil
}

func (d *DirectPool[V]) Stats() Stats {
	return Stats{
		Open:    true,
		NumOpen: int(d.numOpened.Load()),
	}
}

var _ GroupPool[string, io.Closer] = (*DirectGroup[string, io.Closer])(nil)

// DirectGroup 实现 GroupPool 接口，但是每次调用都会返回一个新元素
type DirectGroup[K comparable, V io.Closer] struct {
	Factory GroupFactory[K, V]
}

func (dg *DirectGroup[K, V]) Get(ctx context.Context, key K) (Entry[V], error) {
	v, err := dg.Factory.NewWithKey(ctx, key)
	if err != nil {
		return nil, err
	}
	et := NewOpenEntry(v, dg)
	et.UpdateUsing()
	return et, nil
}

func (dg *DirectGroup[K, V]) Put(e Entry[V], err error) {
	e.Close()
}

func (dg *DirectGroup[K, V]) Close() error {
	return nil
}

func (dg *DirectGroup[K, V]) Stats() Stats {
	return Stats{}
}

func (dg *DirectGroup[K, V]) GroupStats() map[K]Stats {
	return map[K]Stats{}
}
