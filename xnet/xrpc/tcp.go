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

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
	"github.com/xanygo/anygo/xsync"
)

var _ Client = (*TCP)(nil)

type TCP struct {
	Interceptor     []TCPInterceptor
	ServiceRegistry xservice.Registry
	Dialer          xnet.Dialer
}

// dial 拨号
// 即使拨号失败，返回的 *xnet.ConnNode 总是不会为 nil
func (c *TCP) dial(ctx context.Context, downstream xnet.AddrNode, opt xoption.Reader) (*xnet.ConnNode, error) {
	cn := &xnet.ConnNode{
		Addr: downstream,
	}
	var proxyDriver xproxy.Driver
	proxyConfig := xproxy.OptConfig(opt)
	if proxyConfig != nil {
		var err error
		proxyDriver, err = xproxy.Find(proxyConfig.Protocol)
		if err != nil {
			return cn, err
		}
	}

	addr := downstream.Addr
	dialer := c.Dialer
	if dialer == nil {
		dialer = xnet.DefaultDialer
	}
	conn, err := dialer.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		return cn, err
	}
	cn.Conn = conn

	if proxyDriver != nil {
		target := xoption.TargetAddress(opt)
		cn, err = xproxy.Proxy(ctx, proxyDriver, cn, proxyConfig, target)
		if err != nil {
			conn.Close()
			return cn, err
		}
	}

	tc := xoption.GetTLSConfig(opt)
	if tc != nil {
		tlsConn := tls.Client(cn.Conn, tc)
		err = tlsConn.HandshakeContext(ctx)
		if err != nil {
			conn.Close()
			return cn, err
		}
		cn.Conn = tlsConn
		return cn, nil
	}
	return cn, nil
}

func (c *TCP) getServiceRegistry() xservice.Registry {
	if c.ServiceRegistry != nil {
		return c.ServiceRegistry
	}
	return xservice.DefaultRegistry()
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

func (c *TCP) Invoke(ctx context.Context, service string, req Request, resp Response, opts ...Option) (result error) {
	start := time.Now()
	its := c.allITs(ctx)
	for _, it := range its {
		if it.BeforeInvoke == nil {
			continue
		}
		ctx, req, resp, opts = it.BeforeInvoke(ctx, service, req, resp, opts...)
	}

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

	var tryInfo TryInfo
	ser := cfg.ser
	if ser == nil {
		var ok bool
		ser, ok = c.getServiceRegistry().Find(service)
		if !ok {
			result = fmt.Errorf("service %q %w", service, xerror.NotFound)
		}
	}

	// 将临时 option 和 service 的 option 合并
	opt := xoption.Readers(cfg.opt, ser.Option())

	if hr, ok1 := req.(HasOptionReader); ok1 {
		if opt1 := hr.OptionReader(ctx, opt); opt1 != nil {
			opt = xoption.Readers(opt1, cfg.opt, ser.Option())
		}
	}

	ctx = xoption.ContextWithReader(ctx, opt)

	ap := ser.Balancer()
	if ap1 := xbalance.OptReader(opt); ap1 != nil {
		ap = ap1
	}

	tryTotal := xoption.Retry(opt) + 1
	tryInfo = TryInfo{
		Total: tryTotal,
		Start: time.Now(),
	}
	for try := 0; try < tryTotal; try++ {
		tryInfo.Index = try
		var conn *xnet.ConnNode
		conn, result = c.doConnect(ctx, service, ap, opt, cfg, its)

		if result == nil {
			tryInfo.Start = time.Now()
			result = c.doWriteRead(ctx, req, resp, opt, conn)
			tryInfo.End = time.Now()
			for _, it := range its {
				if it.AfterWriteRead != nil {
					it.AfterWriteRead(ctx, service, conn, req, resp, tryInfo, result)
				}
			}
		}
		if result == nil {
			break
		}
	}

	tryInfo.Start = start
	tryInfo.End = time.Now()

	for _, it := range its {
		if it.AfterInvoke != nil {
			it.AfterInvoke(ctx, service, req, resp, tryInfo, result)
		}
	}
	return result
}

func (c *TCP) doConnect(ctx context.Context, service string, ap xbalance.Reader, opt xoption.Reader, cfg *config, its []TCPInterceptor) (*xnet.ConnNode, error) {
	tryTotal := xoption.ConnectRetry(opt) + 1
	connectTimeout := xoption.ConnectTimeout(opt)

	if cfg.ap != nil {
		ap = cfg.ap
	}
	doConnect := func(ctx context.Context, index int) (*xnet.ConnNode, error) {
		tryInfo := TryInfo{
			Total: tryTotal,
			Index: index,
		}
		ctx, cancel := context.WithTimeout(ctx, connectTimeout)
		defer cancel()

		for _, it := range its {
			if it.BeforePickAddress != nil {
				it.BeforePickAddress(ctx, service, tryInfo)
			}
		}

		tryInfo.Start = time.Now()
		node, err := ap.Pick(ctx)
		tryInfo.End = time.Now()

		for _, it := range its {
			if it.AfterPickAddress != nil {
				it.AfterPickAddress(ctx, service, tryInfo, node, err)
			}
		}

		if err != nil {
			return nil, err
		}

		tryInfo.Start = time.Now()
		connNode, err := c.dial(ctx, *node, opt)
		tryInfo.End = time.Now()
		for _, it := range its {
			if it.AfterDial != nil {
				it.AfterDial(ctx, service, tryInfo, connNode, err)
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
	}
	return nil, err
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
	return resp.LoadFrom(ctx, rd, opt)
}

type TCPInterceptor struct {
	BeforeInvoke func(ctx context.Context, service string, req Request, resp Response,
		opts ...Option) (context.Context, Request, Response, []Option)

	BeforePickAddress func(ctx context.Context, service string, try TryInfo)
	AfterPickAddress  func(ctx context.Context, service string, try TryInfo, node *xnet.AddrNode, err error)

	// AfterDial 拨号完成后执行，最多执行 ( retry + 1 ) * ( connectRetry +1) 次
	AfterDial func(ctx context.Context, service string, try TryInfo, conn *xnet.ConnNode, err error)

	// AfterWriteRead 每 Write + Read 完成后都会执行一次，最多执行 retry+1 次
	AfterWriteRead func(ctx context.Context, service string, conn *xnet.ConnNode, req Request, resp Response, try TryInfo, err error)

	// AfterInvoke 在 Invoke 执行完成后，执行一次
	AfterInvoke func(ctx context.Context, service string, req Request, resp Response, try TryInfo, err error)
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

func Invoke(ctx context.Context, service string, req Request, resp Response, opts ...Option) (result error) {
	return DefaultTCPClient().Invoke(ctx, service, req, resp, opts...)
}

var globalTCPInterceptors []TCPInterceptor

func RegisterTCPIT(its ...TCPInterceptor) {
	globalTCPInterceptors = append(globalTCPInterceptors, its...)
}
