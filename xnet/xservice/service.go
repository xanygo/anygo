//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"context"
	"net/http"
	"time"

	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xnaming"
	"github.com/xanygo/anygo/xpp"
)

// Service 下游服务
type Service interface {
	Name() string   // 服务名称
	String() string // 描述信息，可用于调试打印

	Balancer() xbalance.Reader

	Option() Option
}

type Option struct {
	ConnectTimeout time.Duration // 连接超时,可选，默认 1 秒
	ConnectRetry   int           // 每次创建连接时候的重试次数，可选，默认 0。值 >0 时有效
	WriteTimeout   time.Duration // 写超时，可选，默认 1 秒
	ReadTimeout    time.Duration // 读超时，可选，默认 1 秒
	Retry          int           // 重试次数，连接读写任意阶段错误后都会触发重试，可选，默认 0。值 >0 时有效
	HTTP           HTTPOption    // HTTP 协议特有的配置属性，可选
}

func (so *Option) GetConnectTimeout() time.Duration {
	if so.ConnectTimeout > 0 {
		return so.ConnectTimeout
	}
	return time.Second
}

func (so *Option) GetConnectRetry() int {
	if so.ConnectRetry > 0 {
		return so.ConnectRetry
	}
	return 0
}

func (so *Option) GetWriteTimeout() time.Duration {
	if so.WriteTimeout > 0 {
		return so.WriteTimeout
	}
	return time.Second
}

func (so *Option) GetReadTimeout() time.Duration {
	if so.ReadTimeout > 0 {
		return so.ReadTimeout
	}
	return time.Second
}

func (so *Option) GetRetry() int {
	if so.Retry > 0 {
		return so.Retry
	}
	return 0
}

//	func (so *Option) merge(n *Option) {
//		if n.ConnectTimeout > 0 {
//			so.ConnectTimeout = n.ConnectTimeout
//		}
//		if n.ConnectRetry != 0 {
//			so.ConnectRetry = n.ConnectRetry
//		}
//		if n.WriteTimeout > 0 {
//			so.ConnectRetry = n.ConnectRetry
//		}
//		if n.ReadTimeout > 0 {
//			so.ReadTimeout = n.ReadTimeout
//		}
//		if n.Retry != 0 {
//			so.Retry = n.Retry
//		}
//		so.HTTP = so.HTTP.merge(n.HTTP)
//	}
type HTTPOption struct {
	Host   string // 主机名，可选
	HTTPS  bool   // 是否发起 HTTPS 请求，可选，默认 false
	Header http.Header
}

//	func (ho HTTPOption) merge(b HTTPOption) HTTPOption {
//		if b.Host != "" {
//			ho.Host = b.Host
//		}
//		if b.HTTPS {
//			ho.HTTPS = b.HTTPS
//		}
//		for k, vs := range b.Header {
//			ho.Header[k] = vs // 同名 key 覆盖
//		}
//		return ho
//	}
func (ho HTTPOption) Clone() HTTPOption {
	return HTTPOption{
		Host:   ho.Host,
		HTTPS:  ho.HTTPS,
		Header: ho.Header.Clone(),
	}
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	name     string
	balancer xbalance.LoadBalancer
	opt      Option
	nw       *xnaming.Worker
	broker   *xbus.Broker
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

func (ds *serviceImpl) Option() Option {
	return ds.opt
}

var _ xpp.CycleWorker = (*serviceImpl)(nil)

func (ds *serviceImpl) Start(ctx context.Context, cycle time.Duration) error {
	err := xpp.TryStartWorker(ctx, cycle, ds.balancer, ds.nw, ds.broker)
	if err == nil {
		err = ds.balancer.Init(ctx, ds.nw.Nodes())
	}
	return err
}

func (ds *serviceImpl) Stop(ctx context.Context) error {
	ds.broker.Stop()
	return xpp.TryStopWorker(ctx, ds.balancer, ds.nw, ds.broker)
}
