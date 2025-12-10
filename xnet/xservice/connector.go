//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-03

package xservice

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xproxy"
)

var _ xdial.Connector = (*connector)(nil)

type connector struct{}

func (c *connector) Connect(ctx context.Context, addr xnet.AddrNode, opt xoption.Reader) (conn *xnet.ConnNode, err error) {
	if useProxy := xoption.UseProxy(opt); useProxy != "" {
		conn, err = c.connectProxy(ctx, useProxy, addr, opt)
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

func (c *connector) connectProxy(ctx context.Context, proxyName string, target xnet.AddrNode, opt xoption.Reader) (nc *xnet.ConnNode, err error) {
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
		if strings.Contains(proxyName, "://") {
			return c.connectProxyURL(ctx, proxyName, target, opt)
		}
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
	proxyOpt := proxyService.Option()
	conn, err := xdial.Connect(ctx, proxyService.Connector(), *proxyAddr, proxyOpt)
	if err != nil {
		return nil, err
	}

	proxyConn, err := xproxy.Proxy(ctx, proxyDriver, conn, proxyConfig, target, proxyOpt)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return proxyConn, nil
}

// connectProxyURL 处理传入的 UseProxy="http://127.0.0.1:3128" 这种格式
// UseProxy 也支持传入多个代理地址，如 UseProxy="http://127.0.0.1:3128,http://10.10.0.1:3128",若是多个，则每次随机使用一个
func (c *connector) connectProxyURL(ctx context.Context, proxy string, target xnet.AddrNode, opt xoption.Reader) (nc *xnet.ConnNode, err error) {
	items := strings.Split(proxy, ",")
	items = xslice.MapFilter(items, func(index int, item string) (string, bool) {
		item = strings.TrimSpace(item)
		return item, item != ""
	})
	oneProxy := xslice.Rand(items)
	if len(oneProxy) == 0 {
		return nil, fmt.Errorf("invalid proxy option %q", proxy)
	}
	cfg, err := xproxy.ParserProxyURL(oneProxy)
	if err != nil {
		return nil, err
	}
	address := net.JoinHostPort(cfg.Host, cfg.Port)
	proxyAddr := xnet.AddrNode{
		HostPort: address,
		Addr:     xnet.NewAddr("tcp", address),
	}

	conn, err := xdial.Connect(ctx, nil, proxyAddr, opt)
	if err != nil {
		return nil, err
	}
	proxyDriver, err := xproxy.Find(cfg.Protocol)
	if err != nil {
		return nil, fmt.Errorf("proxy %q: %w", proxy, err)
	}

	proxyConn, err := xproxy.Proxy(ctx, proxyDriver, conn, cfg, target, opt)
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

	timeout := xoption.HandshakeTimeout(opt)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	if err = tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("%w, ServerName=%q", err, tc.ServerName)
	}
	conn.AddWrap(tlsConn)
	return conn, nil
}

const Network = "xservice"

// DialerFuncWithServiceName 拨号时，直接使用的是 service 的名字作为 address
func DialerFuncWithServiceName(srv Service) func(ctx context.Context, addr string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		note, err := xbalance.Pick(ctx, srv.Balancer())
		if err != nil {
			return nil, err
		}
		return xdial.Connect(ctx, srv.Connector(), *note, srv.Option())
	}
}

// DialerFuncWithServiceName2 拨号时，直接使用的是 service 的名字作为 address，但是 service 优先从 reg 中查找
func DialerFuncWithServiceName2(reg Registry) func(ctx context.Context, addr string) (net.Conn, error) {
	return func(ctx context.Context, serviceName string) (net.Conn, error) {
		srv, err := FindServiceWithRegistry(reg, serviceName)
		if err != nil {
			return nil, err
		}
		note, err := xbalance.Pick(ctx, srv.Balancer())
		if err != nil {
			return nil, err
		}
		return xdial.Connect(ctx, srv.Connector(), *note, srv.Option())
	}
}
