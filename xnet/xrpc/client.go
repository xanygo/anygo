//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xdial"
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
	WriteTo(ctx context.Context, rw *xnet.ConnNode, opt xoption.Reader) error
}

type Response interface {
	String() string
	LoadFrom(ctx context.Context, req Request, rw *xnet.ConnNode, opt xoption.Reader) error
	xerror.HasErrCode
	xerror.HasErrMsg
	Unwrap() any
}

type HasOptionReader interface {
	OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader
}

type config struct {
	opt       *xoption.Simple
	ap        xbalance.Reader
	service   xservice.Service
	registry  xservice.Registry
	handshake xdial.HandshakeHandler
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

func OptConnectRetry(n int) Option {
	return optionFunc(func(o *config) {
		xoption.SetConnectRetry(o.opt, n)
	})
}

func OptWriteTimeout(t time.Duration) Option {
	return optionFunc(func(o *config) {
		xoption.SetWriteTimeout(o.opt, t)
	})
}

func OptReadTimeout(t time.Duration) Option {
	return optionFunc(func(o *config) {
		xoption.SetReadTimeout(o.opt, t)
	})
}

func OptRetry(n int) Option {
	return optionFunc(func(o *config) {
		xoption.SetRetry(o.opt, n)
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

func OptHandshakeHandler(h xdial.HandshakeHandler) Option {
	return optionFunc(func(o *config) {
		o.handshake = h
	})
}

func OptOptionSetter(fn func(o xoption.Option)) Option {
	return optionFunc(func(o *config) {
		fn(o.opt)
	})
}
