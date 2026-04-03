//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/dsession"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xpolicy"
	"github.com/xanygo/anygo/xnet/xservice"
)

type Client interface {
	// Invoke 发送请求
	// service: 为 string 类型时，是 serviceName。还可以是 xservice.Service 类型，
	Invoke(ctx context.Context, service any, req Request, resp Response, opts ...Option) error
}

type Request interface {
	String() string
	Protocol() string
	APIName() string

	// WriteTo 将请求写入到 w，w 大多数情况下是 *xnet.ConnNode
	WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error
}

type Response interface {
	String() string

	// LoadFrom 从 r  中读取响应数据，r 大多数情况下是 *xnet.ConnNode
	LoadFrom(ctx context.Context, req Request, r io.Reader, opt xoption.Reader) error

	xerror.HasErrCode
	xerror.HasErrMsg
	Unwrap() any
}

type HasOptionReader interface {
	OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader
}

type config struct {
	opt         *xoption.Simple
	ap          xbalance.Reader
	service     xservice.Service
	registry    xservice.Registry
	sessionInit dsession.Starter
}

func (cfg config) getService(srv any) (xservice.Service, error) {
	var serviceName string
	switch sv := srv.(type) {
	case string:
		serviceName = sv
	case xservice.Service:
		return sv, nil
	default:
		return nil, fmt.Errorf("invalid service name: %#v", srv)
	}

	if cfg.service != nil {
		return cfg.service, nil
	}

	if cfg.registry != nil {
		return xservice.FindServiceWithRegistry(cfg.registry, serviceName)
	}
	return xservice.FindService(serviceName)
}

type Option interface {
	withOption(o *config)
}

type optionFunc func(o *config)

func (f optionFunc) withOption(o *config) {
	f(o)
}

func OptConnectTimeout(t time.Duration) Option {
	return optionFunc(func(o *config) {
		xoption.SetConnectTimeout(o.opt, t)
	})
}

// OptConnectRetry 设置单次创建 Socket 超时时间
func OptConnectRetry(n int) Option {
	return optionFunc(func(o *config) {
		xoption.SetConnectRetry(o.opt, n)
	})
}

// OptWriteTimeout 设置单次 Socket 读超时时间
func OptWriteTimeout(t time.Duration) Option {
	return optionFunc(func(o *config) {
		xoption.SetWriteTimeout(o.opt, t)
	})
}

// OptReadTimeout 设置单次 Socket 读超时时间
func OptReadTimeout(t time.Duration) Option {
	return optionFunc(func(o *config) {
		xoption.SetReadTimeout(o.opt, t)
	})
}

// OptTimeout 设置调用 Invoke 方法的整体超时时间
//
//	其他的 OptConnectRetry、OptWriteTimeout、OptReadTimeout 都是针对单次尝试而言的，
//	而调用一此 Invoke 方法，会有 Retry+1 次尝试，实际耗时时间会是 ( Retry + 1 ) * ( ConnectCost + WriteCost + ReadCost )
func OptTimeout(t time.Duration) Option {
	return optionFunc(func(o *config) {
		xoption.SetTimeout(o.opt, t)
	})
}

// OptRetry 设置重置次数为 n，重试策略为每次均重试
func OptRetry(n int) Option {
	return OptRetryWithPolicy(n, xpolicy.AlwaysRetry())
}

// OptRetryPolicy 单独设置重试策略
func OptRetryPolicy(policy *xpolicy.Retry) Option {
	return optionFunc(func(o *config) {
		xoption.SetRetryPolicy(o.opt, policy)
	})
}

// OptRetryWithPolicy 设置重试次数和重试策略，若 policy 为 nil则使用默认策略
func OptRetryWithPolicy(retry int, policy *xpolicy.Retry) Option {
	return optionFunc(func(o *config) {
		xoption.SetRetry(o.opt, retry)
		xoption.SetRetryPolicy(o.opt, policy)
	})
}

func OptAddr(addr ...net.Addr) Option {
	return optionFunc(func(o *config) {
		o.ap = xbalance.NewStaticByAddr(addr...)
	})
}

func OptHostPort(hostPort string) Option {
	return OptAddr(xnet.NewAddr("tcp", hostPort))
}

func OptTLSConfig(c *tls.Config) Option {
	return optionFunc(func(o *config) {
		xoption.SetTLSConfig(o.opt, c)
	})
}

func OptService(s xservice.Service) Option {
	return optionFunc(func(o *config) {
		o.service = s
	})
}

func OptServiceRegistry(s xservice.Registry) Option {
	return optionFunc(func(o *config) {
		o.registry = s
	})
}

// OptSessionInit 注册在创建连接后，开始创建会话的逻辑。如身份认证（Redis）、协议升级（HTTP Upgrade）等
//
// 局限性：
//
//  1. 对于不使用连接池的短连接，由于每次请求都是创建新连接，无使用限制
//  2. 对于使用了连接池的长连接，建议在 service 配置文件中配置，否则容易出现不一致：传入不同的 Starter 逻辑，导致会话信息混乱
//  3. 使用此方法后，service 配置里的 SessionInit 里的配置段落将不生效
func OptSessionInit(h dsession.Starter) Option {
	return optionFunc(func(o *config) {
		o.sessionInit = h
	})
}

func OptOptionSetter(fn func(o xoption.Option)) Option {
	return optionFunc(func(o *config) {
		fn(o.opt)
	})
}

// OptUseProxy 使用代理，
//
// proxyURL 可以是:
//
//	1.一个 有 Proxy 配置段落的 service 的名字
//	2. http 或者  https 代理的地址，如 http://127.0.0.1:3128
func OptUseProxy(proxyURL string) Option {
	return optionFunc(func(o *config) {
		xoption.SetUseProxy(o.opt, proxyURL)
	})
}
