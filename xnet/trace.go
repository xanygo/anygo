//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xnet

import (
	"context"
	"net"
	"time"

	"github.com/xanygo/anygo/xlog"
)

const UserAgent = "anygo-xrpc/1.0"

// PrintLogITs 打印域名解析过程、拨号过程日志的拦截器
var PrintLogITs = []Interceptor{
	PrintResolverLogIT,
	PrintDialLogIT,
}

// PrintResolverLogIT 打印域名解析过程日志的拦截器
var PrintResolverLogIT = &ResolverInterceptor{
	LookupIP: func(ctx context.Context, network, host string, invoker LookupIPFunc) ([]net.IP, error) {
		start := time.Now()
		result, err := invoker(ctx, network, host)
		cost := time.Since(start)
		attrs := []xlog.Attr{
			xlog.String("network", network),
			xlog.String("host", host),
			xlog.String("cost", cost.String()),
			xlog.Any("result", result),
			xlog.ErrorAttr("error", err),
		}
		if err == nil {
			xlog.Info(ctx, "LookupIP", attrs...)
		} else {
			xlog.Error(ctx, "LookupIP", attrs...)
		}
		return result, err
	},
}

// PrintDialLogIT 打印拨号过程日志的拦截器
var PrintDialLogIT = &DialerInterceptor{
	DialContext: func(ctx context.Context, network string, address string, invoker DialContextFunc) (net.Conn, error) {
		start := time.Now()
		result, err := invoker(ctx, network, address)
		cost := time.Since(start)
		attrs := []xlog.Attr{
			xlog.String("network", network),
			xlog.String("address", address),
			xlog.Any("result", result),
			xlog.ErrorAttr("error", err),
			xlog.String("cost", cost.String()),
			xlog.ErrorAttr("error", err),
		}
		if err == nil {
			xlog.Info(ctx, "DialContext", attrs...)
		} else {
			xlog.Error(ctx, "DialContext", attrs...)
		}

		return result, err
	},
}
