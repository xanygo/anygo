//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

type Client interface {
	Invoke(ctx context.Context, service string, req Request, resp Response, opts ...Option) error
}

type Request interface {
	String() string
	Protocol() string
	APIName() string
	WriteTo(ctx context.Context, w *xnet.ConnNode, opt xoption.Reader) error
}

type Response interface {
	String() string
	LoadFrom(ctx context.Context, r io.Reader, opt xoption.Reader) error
	xerror.HasErrCode
	xerror.HasErrMsg
	Unwrap() any
}

type HasOptionReader interface {
	OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader
}

type config struct {
	opt *xoption.Simple
	ap  xbalance.Reader
	ser xservice.Service
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
		o.ser = s
	})
}
