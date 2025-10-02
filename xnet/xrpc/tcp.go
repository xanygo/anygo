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
	"slices"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ Client = (*TCP)(nil)

type TCP struct {
	Interceptor     []TCPInterceptor
	ServiceRegistry xservice.Registry
	Dialer          xnet.Dialer
}

func (c *TCP) getDialer() xnet.Dialer {
	if c.Dialer != nil {
		return c.Dialer
	}
	return xnet.DefaultDialer
}

func (c *TCP) allITs(ctx context.Context) []TCPInterceptor {
	its := slices.Clone(globalTCPInterceptors)
	if len(c.Interceptor) > 0 {
		its = append(its, c.Interceptor...)
	}
	if items := TCPITFromContext(ctx); len(items) > 0 {
		its = append(its, items...)
	}
	return its
}

func (c *TCP) getServiceName(srv any) string {
	switch sv := srv.(type) {
	case string:
		return sv
	case xservice.Service:
		return sv.Name()
	default:
		return fmt.Sprintf("invalid-%v", sv)
	}
}

func (c *TCP) Invoke(ctx context.Context, srv any, req Request, resp Response, opts ...Option) (result error) {
	action := NewAction("Invoke", 0)
	its := c.allITs(ctx)

	cfg := &config{
		opt: xoption.NewSimple(),
	}
	ctxOpts := OptionsFromContext(ctx)
	for _, opt := range ctxOpts {
		opt.withOption(cfg)
	}

	for _, o := range opts {
		o.withOption(cfg)
	}

	serviceName := c.getServiceName(srv)

	service, err := cfg.getService(srv, c.ServiceRegistry)

	defer func() {
		action.End = time.Now()
		for _, it := range its {
			if it.AfterInvoke != nil {
				it.AfterInvoke(ctx, serviceName, req, resp, action, result)
			}
		}
	}()

	for _, it := range its {
		if it.BeforeInvoke == nil {
			continue
		}
		ctx, req, resp, opts = it.BeforeInvoke(ctx, serviceName, req, resp, opts...)
	}

	if err != nil {
		result = err
		return err
	}

	// 将临时 option 和 service 的 option 合并
	opt := xoption.Readers(cfg.opt, service.Option())

	if hr, ok1 := req.(HasOptionReader); ok1 {
		if opt1 := hr.OptionReader(ctx, opt); opt1 != nil {
			opt = xoption.Readers(opt1, cfg.opt, service.Option())
		}
	}

	ctx = xoption.ContextWithReader(ctx, opt)

	td := tcpClientDialer{
		dialer:      c.getDialer(),
		registry:    cfg.getRegistry(c.ServiceRegistry),
		interceptor: its,
		service:     service,
		option:      opt,
		config:      cfg,
	}
	connectFn, err := td.getConnectFunc()
	if err != nil {
		return err
	}

	tryTotal := xoption.Retry(opt) + 1

	for try := 0; try < tryTotal; try++ {
		var conn *xnet.ConnNode
		conn, result = connectFn(ctx)

		if result == nil {
			actionWR := NewAction("WriteRead", tryTotal)
			actionWR.TryIndex = try
			result = c.doWriteRead(ctx, req, resp, opt, conn)
			actionWR.End = time.Now()
			for _, it := range its {
				if it.AfterWriteRead != nil {
					it.AfterWriteRead(ctx, serviceName, conn, req, resp, actionWR, result)
				}
			}
		}
		if result == nil {
			return nil
		}
		if err1 := ctx.Err(); err1 != nil {
			break
		}
	}

	return result
}

func (c *TCP) doWriteRead(ctx context.Context, req Request, resp Response, opt xoption.Reader, node *xnet.ConnNode) (err error) {
	var conn net.Conn = node.Conn
	defer conn.Close()
	totalTimeout := xoption.WriteReadTimeout(opt)
	ctx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()
	if err = conn.SetDeadline(time.Now().Add(totalTimeout)); err != nil {
		return err
	}
	start := time.Now()
	// 暂时不将读写超时分开控制
	err = req.WriteTo(ctx, node, opt)
	if err != nil {
		return err
	}
	if node.TlsConn != nil {
		conn = node.TlsConn
	}
	maxBody := xoption.MaxResponseSize(opt)
	rd := io.LimitReader(conn, maxBody)
	err = resp.LoadFrom(ctx, req, rd, opt)
	if err != nil {
		return fmt.Errorf("%w, cost=%s", err, time.Since(start).String())
	}
	return err
}

type tcpClientDialer struct {
	dialer      xnet.Dialer
	registry    xservice.Registry
	interceptor []TCPInterceptor
	service     xservice.Service
	option      xoption.Reader
	config      *config
}

func (td tcpClientDialer) getConnectFunc() (func(ctx context.Context) (*xnet.ConnNode, error), error) {
	// 关于命名：
	// downstream: 用于读取即将连接的服务器地址，包括代理服务器地址
	// target:  最终目标服务器地址，被代理的服务器地址
	// 在有代理的情况下，downstream 是代理服务器地址，target 是被代理服务器地址

	downstreamAP := td.service.Balancer()
	targetAP := downstreamAP

	if tp := xbalance.OptTarget(td.option); tp != nil {
		targetAP = tp
	}

	connectServiceName := td.service.Name()
	option := td.option

	proxyConfig := xproxy.OptConfig(td.option)
	var proxyService xservice.Service
	if proxyConfig != nil && proxyConfig.Use != "" {
		var ok bool
		proxyService, ok = td.registry.Find(proxyConfig.Use)
		if !ok {
			return nil, fmt.Errorf("proxy service %q %w", proxyConfig.Use, xerror.NotFound)
		}
		opt := proxyService.Option()
		pc := xproxy.OptConfig(opt)
		if pc == nil || pc.Protocol == "" {
			return nil, fmt.Errorf("serivce %q missing Proxy config", proxyConfig.Use)
		}
		connectServiceName = proxyConfig.Use
		downstreamAP = proxyService.Balancer()
		option = xoption.Readers(td.config.opt, opt)
	}

	if ap1 := xbalance.OptDownstream(option); ap1 != nil {
		downstreamAP = ap1
	}

	if td.config.ap != nil {
		downstreamAP = td.config.ap
	}

	return func(ctx context.Context) (*xnet.ConnNode, error) {
		return td.doConnect(ctx, connectServiceName, downstreamAP, option, targetAP, td.option)
	}, nil
}

func (td tcpClientDialer) doConnect(ctx context.Context, service string, downstreamAP xbalance.Reader, downstreamOpt xoption.Reader, targetAP xbalance.Reader, targetOpt xoption.Reader) (*xnet.ConnNode, error) {
	ctx = xoption.ContextWithReader(ctx, downstreamOpt)

	tryTotal := xoption.ConnectRetry(downstreamOpt) + 1
	connectTimeout := xoption.ConnectTimeout(downstreamOpt)

	doConnect := func(ctx context.Context, index int) (*xnet.ConnNode, error) {
		ctx1, cancel := context.WithTimeout(ctx, connectTimeout)
		defer cancel()

		action := NewAction("Pick", tryTotal)
		action.TryIndex = index
		for _, it := range td.interceptor {
			if it.BeforePickAddress != nil {
				it.BeforePickAddress(ctx1, service, action)
			}
		}

		node, err := downstreamAP.Pick(ctx1)
		action.End = time.Now()

		for _, it := range td.interceptor {
			if it.AfterPickAddress != nil {
				it.AfterPickAddress(ctx1, service, action, node, err)
			}
		}
		if err != nil {
			return nil, err
		}
		targetNode := node
		if targetAP != nil && targetAP != downstreamAP {
			ta, err := targetAP.Pick(ctx1)
			if err != nil {
				return nil, err
			}
			targetNode = ta
		}

		var connNode *xnet.ConnNode
		action = NewAction("Dial", tryTotal)
		action.TryIndex = index
		connNode, err = td.dial(ctx1, *node, downstreamOpt, *targetNode, targetOpt)
		action.End = time.Now()
		if err != nil && connNode == nil {
			connNode = &xnet.ConnNode{
				Addr: *node,
			}
		}
		for _, it := range td.interceptor {
			if it.AfterDial != nil {
				it.AfterDial(ctx1, service, action, connNode, err)
			}
		}

		return connNode, err
	}

	var err error
	var conn *xnet.ConnNode
	for i := 0; i < tryTotal; i++ {
		if conn, err = doConnect(ctx, i); err == nil {
			return conn, nil
		}
		if err1 := ctx.Err(); err1 != nil {
			break
		}
	}
	return conn, err
}

func (td tcpClientDialer) dial(ctx context.Context, downstream xnet.AddrNode, downstreamOpt xoption.Reader, target xnet.AddrNode, targetOpt xoption.Reader) (*xnet.ConnNode, error) {
	var proxyDriver xproxy.Driver
	proxyConfig := xproxy.OptConfig(downstreamOpt)
	if proxyConfig != nil {
		var err error
		proxyDriver, err = xproxy.Find(proxyConfig.Protocol)
		if err != nil {
			return nil, err
		}
	}

	addr := downstream.Addr
	conn, err := td.dialer.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	result := &xnet.ConnNode{
		Addr: downstream,
		Conn: conn,
	}

	if proxyDriver != nil {
		result, err = xproxy.Proxy(ctx, proxyDriver, result, proxyConfig, target)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	result.Conn.SetDeadline(time.Now().Add(xoption.HandshakeTimeout(targetOpt)))
	defer result.Conn.SetDeadline(time.Time{})

	// TLS 握手
	result, err = td.targetTLSHandshake(ctx, result, targetOpt, target)
	if err != nil {
		return result, err
	}

	// 协议层面的握手，如 redis 需要发送 HELLO request
	if td.config.handshake != nil {
		ret, err1 := td.config.handshake.Handshake(ctx, result, targetOpt)
		if err1 != nil {
			_ = result.Conn.Close()
			return result, err1
		}
		result.Handshake = ret
	}
	return result, err
}

func (td tcpClientDialer) targetTLSHandshake(ctx context.Context, conn *xnet.ConnNode, targetOpt xoption.Reader, target xnet.AddrNode) (*xnet.ConnNode, error) {
	tc := xoption.GetTLSConfig(targetOpt)
	if tc == nil {
		return conn, nil
	}
	tc = tc.Clone()
	if tc.ServerName == "" {
		tc.ServerName = target.Host()
	}
	tlsConn := tls.Client(conn.Conn, tc)

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		conn.Conn.Close()
		return nil, fmt.Errorf("%w, ServerName=%q", err, tc.ServerName)
	}
	conn.Conn = tlsConn
	return conn, nil
}

type TCPInterceptor struct {
	BeforeInvoke func(ctx context.Context, service string, req Request, resp Response,
		opts ...Option) (context.Context, Request, Response, []Option)

	BeforePickAddress func(ctx context.Context, service string, try Action)
	AfterPickAddress  func(ctx context.Context, service string, try Action, node *xnet.AddrNode, err error)

	// AfterDial 拨号完成后执行，最多执行 ( retry + 1 ) * ( connectRetry +1) 次
	AfterDial func(ctx context.Context, service string, try Action, conn *xnet.ConnNode, err error)

	// AfterWriteRead 每 Write + Read 完成后都会执行一次，最多执行 retry+1 次
	AfterWriteRead func(ctx context.Context, service string, conn *xnet.ConnNode, req Request, resp Response, try Action, err error)

	// AfterInvoke 在 Invoke 执行完成后，执行一次
	AfterInvoke func(ctx context.Context, service string, req Request, resp Response, try Action, err error)
}

var defaultTCPClient = &xsync.OnceInit[*TCP]{
	New: func() *TCP {
		return &TCP{}
	},
}

func DefaultTCPClient() *TCP {
	return defaultTCPClient.Load()
}

func SetDefaultTCPClient(c *TCP) {
	defaultTCPClient.Store(c)
}

func Invoke(ctx context.Context, service any, req Request, resp Response, opts ...Option) (result error) {
	return DefaultTCPClient().Invoke(ctx, service, req, resp, opts...)
}

var globalTCPInterceptors []TCPInterceptor

func RegisterTCPIT(its ...TCPInterceptor) {
	globalTCPInterceptors = append(globalTCPInterceptors, its...)
}
