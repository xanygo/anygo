//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xservice

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xoption"
)

var _ xdial.Connector = (*connector)(nil)

type connector struct{}

func (c *connector) Connect(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (conn *xnet.ConnNode, err error) {
	if useProxy := xoption.UseProxy(opt); useProxy != "" {
		conn, err = c.connectProxy(ctx, useProxy, addr)
	} else {
		conn, err = xdial.DefaultConnector().Connect(ctx, addr, opt)
	}
	if err != nil {
		return conn, err
	}

	// TLS 握手
	conn, err = c.tlsHandshake(ctx, conn, opt, addr)
	if err != nil {
		return conn, err
	}

	protocol := xoption.Protocol(opt)
	if protocol != "" {
		ret, err1 := xdial.Handshake(ctx, protocol, conn, opt)
		if err1 != nil {
			_ = conn.NetConn().Close()
			return conn, err
		}
		conn.Handshake = ret
	}

	return conn, nil
}

func (c *connector) connectProxy(ctx context.Context, proxyName string, target xnet.AddrNode) (*xnet.ConnNode, error) {
	proxyService, err := FindService(proxyName)
	if err != nil {
		return nil, err
	}
	proxyConfig := xproxy.OptConfig(proxyService.Option())
	if proxyConfig == nil {
		return nil, fmt.Errorf("service %q missing Proxy option", proxyName)
	}
	proxyDriver, err := xproxy.Find(proxyConfig.Protocol)
	if err != nil {
		return nil, fmt.Errorf("proxy %q: %w", proxyName, err)
	}

	proxyAddr, err := proxyService.Balancer().Pick(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := proxyService.Connector().Connect(ctx, *proxyAddr, proxyService.Option())
	if err != nil {
		return conn, err
	}

	proxyConn, err := xproxy.Proxy(ctx, proxyDriver, conn, proxyConfig, target)
	if err != nil {
		_ = conn.Close()
		return conn, err
	}
	return proxyConn, nil
}

func (c *connector) tlsHandshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader, target xnet.AddrNode) (*xnet.ConnNode, error) {
	tc := xoption.GetTLSConfig(opt)
	if tc == nil {
		return conn, nil
	}
	tc = tc.Clone()
	if tc.ServerName == "" {
		tc.ServerName = target.Host()
	}
	tlsConn := tls.Client(conn.NetConn(), tc)

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Conn.Close()
		return conn, fmt.Errorf("%w, ServerName=%q", err, tc.ServerName)
	}
	conn.TlsConn = tlsConn
	return conn, nil
}
