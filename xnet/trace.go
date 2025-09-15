//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xnet

import (
	"context"
	"log"
	"net"
	"time"
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
		log.Printf("LookupIP( %q, %q ) = ( %v, %v ) cost = %s\n", network, host, result, err, cost)
		return result, err
	},
}

// PrintDialLogIT 打印拨号过程日志的拦截器
var PrintDialLogIT = &DialerInterceptor{
	DialContext: func(ctx context.Context, network string, address string, invoker DialContextFunc) (net.Conn, error) {
		start := time.Now()
		result, err := invoker(ctx, network, address)
		cost := time.Since(start)
		if err == nil {
			log.Printf("DialContext( %q, %q ), localAddr=%q, cost = %s\n", network, address, result.LocalAddr().String(), cost)
		} else {
			log.Printf("DialContext( %q, %q ) failed, err=%v, cost = %s\n", network, address, err, cost)
		}
		return result, err
	},
}
