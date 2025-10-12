//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xservice

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
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
		conn, err = xdial.Connect(ctx, nil, addr, opt)
	}
	if err != nil {
		return conn, err
	}
	inConn := conn

	conn, err = c.tlsHandshake(ctx, inConn, opt, addr)
	if err != nil {
		inConn.Close()
		return nil, err
	}

	protocol := xoption.Protocol(opt)
	if protocol != "" {
		ret, err1 := xdial.Handshake(ctx, protocol, conn, opt)
		if err1 != nil {
			_ = conn.Close()
			return nil, err1
		}
		conn.Handshake = ret
	}

	return conn, nil
}

func (c *connector) connectProxy(ctx context.Context, proxyName string, target xnet.AddrNode) (nc *xnet.ConnNode, err error) {
	ctx, span := xmetric.Start(ctx, "ConnectProxy")
	defer func() {
		span.RecordError(err)
		span.End()
	}()
	span.SetAttributes(
		xmetric.AnyAttr("UseProxy", proxyName),
		xmetric.AnyAttr("Target", target.HostPort),
	)

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

	proxyAddr, err := xbalance.Pick(ctx, proxyService.Balancer())

	if err != nil {
		return nil, err
	}

	conn, err := xdial.Connect(ctx, proxyService.Connector(), *proxyAddr, proxyService.Option())
	if err != nil {
		return nil, err
	}

	proxyConn, err := xproxy.Proxy(ctx, proxyDriver, conn, proxyConfig, target)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return proxyConn, nil
}

func (c *connector) tlsHandshake(ctx context.Context, conn *xnet.ConnNode, opt xoption.Reader, target xnet.AddrNode) (nc *xnet.ConnNode, err error) {
	tc := xoption.GetTLSConfig(opt)
	if tc == nil {
		return conn, nil
	}
	ctx, span := xmetric.Start(ctx, "TLSHandshake")
	defer func() {
		span.RecordError(err)
		span.End()
	}()
	tc = tc.Clone()
	if tc.ServerName == "" {
		tc.ServerName = target.Host()
	}
	if tc.MinVersion == 0 {
		tc.MinVersion = tls.VersionTLS12
	}
	span.SetAttributes(
		xmetric.AnyAttr("ServerName", tc.ServerName),
		xmetric.AnyAttr("SkipVerify", tc.InsecureSkipVerify),
	)
	tlsConn := tls.Client(conn.Outer(), tc)

	if err = tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("%w, ServerName=%q", err, tc.ServerName)
	}
	conn.AddWrap(tlsConn)
	return conn, nil
}
