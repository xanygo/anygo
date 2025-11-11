//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"context"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xnaming"
	"github.com/xanygo/anygo/xpp"
)

// Service 下游服务
type Service interface {
	Name() string   // 服务名称
	String() string // 描述信息，可用于调试打印

	Balancer() xbalance.Reader      // 负载均衡器
	Connector() xdial.Connector     // 拨号器，包括拨号和握手逻辑
	GroupPool() xdial.ConnGroupPool // 网络连接池

	Option() xoption.Reader

	xpp.Worker
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	name      string
	balancer  xbalance.LoadBalancer
	connector xdial.Connector
	opt       *xoption.Dynamic
	nw        *xnaming.Worker
	broker    *xbus.Broker
	pool      xdial.ConnGroupPool
}

func (ds *serviceImpl) Name() string {
	return ds.name
}

func (ds *serviceImpl) String() string {
	return ds.name
}

func (ds *serviceImpl) Balancer() xbalance.Reader {
	return ds.balancer
}

func (ds *serviceImpl) Connector() xdial.Connector {
	return ds.connector
}

func (ds *serviceImpl) Option() xoption.Reader {
	return ds.opt
}

func (ds *serviceImpl) GroupPool() xdial.ConnGroupPool {
	return ds.pool
}

func (ds *serviceImpl) Start(ctx context.Context) error {
	err := xpp.TryStartWorker(ctx, ds.balancer, ds.nw, ds.broker)
	if err == nil {
		err = ds.balancer.Init(ctx, ds.nw.Nodes())
	}
	return err
}

func (ds *serviceImpl) Stop(ctx context.Context) error {
	ds.broker.Stop()
	return xpp.TryStopWorker(ctx, ds.balancer, ds.nw, ds.broker)
}
