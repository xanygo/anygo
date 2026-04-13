//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-13

package xproxy

import (
	"context"
	"net"
	"net/url"

	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/internal"
)

// NewDialer 传入代理的 url 地址，返回通过代理的拨号方法
func NewDialer(rawURL string) (xnet.Dialer, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return NewDialerFromURL(u)
}

func NewDialerFromURL(proxy *url.URL) (xnet.Dialer, error) {
	d, err0 := Find(proxy.Scheme)
	if err0 != nil {
		return nil, err0
	}
	host := proxy.Hostname()
	port := proxy.Port()
	if port == "" {
		port, err0 = internal.SchemePort(proxy.Scheme)
		if err0 != nil {
			return nil, err0
		}
	}
	proxyAddress := net.JoinHostPort(host, port)
	proxyAddrNote := xnet.AddrNode{
		HostPort: proxyAddress,
		Addr:     internal.NewAddr("tcp", proxyAddress),
	}
	cfg := &Config{
		Protocol: proxy.Scheme,
	}
	if u := proxy.User; u != nil {
		cfg.Username = u.Username()
		cfg.Password, _ = u.Password()
	}

	return xnet.DialContextFunc(func(ctx context.Context, network, address string) (net.Conn, error) {
		target := xnet.AddrNode{
			HostPort: address,
			Addr:     internal.NewAddr(network, address),
		}
		conn, err := xnet.DialContext(ctx, "tcp", proxyAddress)
		if err != nil {
			return nil, err
		}
		cn := &xnet.ConnNode{
			Conn: conn,
			Addr: proxyAddrNote,
		}
		return Proxy(ctx, d, cn, cfg, target, nil)
	}), nil
}
