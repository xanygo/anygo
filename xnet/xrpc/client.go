//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet/xservice"
)

type Client interface {
	Invoke(ctx context.Context, service string, req Request, resp Response, opts ...Option) error
}

type Request interface {
	String() string
	APIName() string
	WriteTo(ctx context.Context, w net.Conn, opt *xservice.Option) error
}

type Response interface {
	String() string
	io.ReaderFrom
	xerror.HasErrCode
	xerror.HasErrMsg
}

type Option interface {
	withOption(o *xservice.Option)
}

type optionFunc func(o *xservice.Option)

func (f optionFunc) withOption(o *xservice.Option) {
	f(o)
}

func OptConnectTimeout(t time.Duration) Option {
	return optionFunc(func(o *xservice.Option) {
		o.ConnectTimeout = t
	})
}

func OptConnectRetry(n int) Option {
	return optionFunc(func(o *xservice.Option) {
		o.ConnectRetry = n
	})
}

func OptWriteTimeout(t time.Duration) Option {
	return optionFunc(func(o *xservice.Option) {
		o.WriteTimeout = t
	})
}

func OptReadTimeout(t time.Duration) Option {
	return optionFunc(func(o *xservice.Option) {
		o.ReadTimeout = t
	})
}

func OptRetry(n int) Option {
	return optionFunc(func(o *xservice.Option) {
		o.Retry = n
	})
}

func OptAddr(addr net.Addr) Option {
	return optionFunc(func(o *xservice.Option) {})
}
