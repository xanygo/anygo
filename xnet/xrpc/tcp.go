//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xnaming"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ Client = (*TCPClient)(nil)

type TCPClient struct {
	Interceptor     []TCPClientInterceptor
	ServiceRegistry xservice.Registry
	Dialer          xnet.Dialer
}

func (c *TCPClient) dial(ctx context.Context, addr net.Addr, opt xoption.Reader) (net.Conn, error) {
	dialer := c.Dialer
	if dialer == nil {
		dialer = xnet.DefaultDialer
	}
	return dialer.DialContext(ctx, addr.Network(), addr.String())
}

func (c *TCPClient) getServiceRegistry() xservice.Registry {
	if c.ServiceRegistry != nil {
		return c.ServiceRegistry
	}
	return xservice.DefaultRegistry()
}

func (c *TCPClient) Invoke(ctx context.Context, service string, req Request, resp Response, opts ...Option) (result error) {
	var err1 error
	for i := 0; i < len(c.Interceptor); i++ {
		it := c.Interceptor[i]
		if it.BeforeInvoke != nil {
			ctx, service, req, resp, opts, err1 = it.BeforeInvoke(ctx, service, req, resp, opts...)
			if err1 != nil {
				result = err1
			}
		}
	}
	cfg := &config{
		opt: xoption.NewMapOption(),
	}
	for _, o := range opts {
		o.withOption(cfg)
	}

	if result == nil {
		ser, ok := c.getServiceRegistry().Find(service)
		if !ok {
			result = fmt.Errorf("service %q %w", service, xerror.NotFound)
		} else {
			// 将临时 option 和 service 的 option 合并
			opt := xoption.Readers(cfg.opt, ser.Option())

			ctx = xoption.ContextWithReader(ctx, opt)

			ap := ser.Balancer()
			if hb, ok1 := req.(xbalance.HasReader); ok1 {
				if ap1 := hb.Balancer(); ap1 != nil {
					ap = ap1
				}
			}

			tryTotal := xoption.Retry(opt) + 1
			for try := 0; try < tryTotal; try++ {
				var conn net.Conn
				conn, result = c.doConnect(ctx, service, ap, opt, cfg)
				if result == nil {
					result = c.doWriteRead(ctx, req, resp, opt, conn)
				}
				if result == nil {
					break
				}
			}
		}
	}

	for i := 0; i < len(c.Interceptor); i++ {
		it := c.Interceptor[i]
		if it.AfterInvoke != nil {
			result = it.AfterInvoke(ctx, service, req, resp, result)
		}
	}
	return result
}

func (c *TCPClient) doConnect(ctx context.Context, service string, ap xbalance.Reader, opt xoption.Reader, cfg *config) (net.Conn, error) {
	tryTotal := xoption.ConnectRetry(opt) + 1
	connectTimeout := xoption.ConnectTimeout(opt)

	if cfg.ap != nil {
		ap = cfg.ap
	}

	doConnect := func(ctx context.Context, index int) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(ctx, connectTimeout)
		defer cancel()

		node, err := ap.Pick(ctx)
		for i := 0; i < len(c.Interceptor); i++ {
			it := c.Interceptor[i]
			if it.AfterPickAddress != nil {
				it.AfterPickAddress(ctx, service, node, err)
			}
		}

		if err != nil {
			return nil, err
		}
		return c.dial(ctx, node.Addr(), opt)
	}

	var err error
	var conn net.Conn
	for i := 0; i < tryTotal; i++ {
		if conn, err = doConnect(ctx, i); err == nil {
			return conn, nil
		}
	}
	return nil, err
}

func (c *TCPClient) doWriteRead(ctx context.Context, req Request, resp Response, opt xoption.Reader, conn net.Conn) (err error) {
	defer conn.Close()
	totalTimeout := xoption.WriteReadTimeout(opt)
	ctx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()
	if err = conn.SetDeadline(time.Now().Add(totalTimeout)); err != nil {
		return err
	}
	// 暂时不将读写超时分开控制
	err = req.WriteTo(ctx, conn, opt)
	if err != nil {
		return err
	}
	maxBody := xoption.MaxResponseSize(opt)
	rd := io.LimitReader(conn, maxBody)
	return resp.LoadFrom(ctx, rd, opt)
}

type TCPClientInterceptor struct {
	BeforeInvoke func(ctx context.Context, service string, req Request, resp Response,
		opts ...Option) (context.Context, string, Request, Response, []Option, error)

	BeforePickAddress func(ctx context.Context, service string)
	AfterPickAddress  func(ctx context.Context, service string, node xnaming.Node, err error)

	AfterInvoke func(ctx context.Context, service string, req Request, resp Response, err error) error
}

var defaultTCPClient = &TCPClient{}

func Invoke(ctx context.Context, service string, req Request, resp Response, opts ...Option) (result error) {
	return defaultTCPClient.Invoke(ctx, service, req, resp, opts...)
}
