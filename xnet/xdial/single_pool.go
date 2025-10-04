//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xdial

import (
	"context"
	"sync/atomic"

	"github.com/xanygo/anygo/ds/xpool"
	"github.com/xanygo/anygo/xnet"
)

var _ ConnPool = (*SinglePool)(nil)

type SinglePool struct {
	Addr      xnet.AddrNode
	Connector Connector
	Option    *xpool.Option

	pool xpool.Pool[*xnet.ConnNode]
}

func (s *SinglePool) Init() {
	fac := &ConnFactory{
		Single: s.factory,
	}
	s.pool = xpool.New[*xnet.ConnNode](s.Option, fac)
}

func (s *SinglePool) factory(ctx context.Context) (*xnet.ConnNode, error) {
	return Connect(ctx, s.Connector, s.Addr, nil)
}

func (s *SinglePool) Get(ctx context.Context) (xpool.Entry[*xnet.ConnNode], error) {
	return s.pool.Get(ctx)
}

func (s *SinglePool) Put(e xpool.Entry[*xnet.ConnNode], err error) {
	s.pool.Put(e, err)
}

func (s *SinglePool) Close() error {
	return s.pool.Close()
}

func (s *SinglePool) Stats() xpool.Stats {
	return s.pool.Stats()
}

var _ ConnPool = (*SingleDirectPool)(nil)

type SingleDirectPool struct {
	Addr      xnet.AddrNode
	Connector Connector
	numOpened atomic.Int64
}

func (np *SingleDirectPool) Get(ctx context.Context) (xpool.Entry[*xnet.ConnNode], error) {
	conn, err := Connect(ctx, np.Connector, np.Addr, nil)
	if err != nil {
		return nil, err
	}
	np.numOpened.Add(1)
	e := xpool.NewOpenEntry[*xnet.ConnNode](conn, np)
	e.UpdateUsing()
	return e, nil
}

func (np *SingleDirectPool) Put(e xpool.Entry[*xnet.ConnNode], err error) {
	e.Close()
	np.numOpened.Add(-1)
}

func (np *SingleDirectPool) Close() error {
	return nil
}

func (np *SingleDirectPool) Stats() xpool.Stats {
	return xpool.Stats{
		Open:    true,
		NumOpen: int(np.numOpened.Load()),
	}
}
