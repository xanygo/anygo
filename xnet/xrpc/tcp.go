//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xnaming"
	xservice2 "github.com/xanygo/anygo/xnet/xservice"
)

var _ Client = (*TCPClient)(nil)

type TCPClient struct {
	Interceptor     []TCPClientInterceptor
	ServiceRegistry xservice2.Registry
	Dialer          xnet.Dialer
}

func (c *TCPClient) dial(ctx context.Context, addr net.Addr, opt *xservice2.Option) (net.Conn, error) {
	dialer := c.Dialer
	if dialer == nil {
		dialer = xnet.DefaultDialer
	}
	return dialer.DialContext(ctx, addr.Network(), addr.String())
}

func (c *TCPClient) getServiceRegistry() xservice2.Registry {
	if c.ServiceRegistry != nil {
		return c.ServiceRegistry
	}
	return xservice2.DefaultRegistry()
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
	opt := &xservice2.Option{}
	for _, o := range opts {
		o.withOption(opt)
	}

	if result == nil {
		ser, ok := c.getServiceRegistry().Find(service)
		if !ok {
			result = fmt.Errorf("service %q %w", service, xerror.NotFound)
		} else {
			for i := 0; i < len(c.Interceptor); i++ {
				it := c.Interceptor[i]
				if it.BeforePickAddress != nil {
					it.BeforePickAddress(ctx, service)
				}
			}
			var node xnaming.Node
			node, result = ser.Balancer().Pick(ctx)
			for i := 0; i < len(c.Interceptor); i++ {
				it := c.Interceptor[i]
				if it.AfterPickAddress != nil {
					it.AfterPickAddress(ctx, service, node, result)
				}
			}
			var conn net.Conn
			if result == nil {
				conn, result = c.dial(ctx, node.Addr(), opt)
			}
			if result == nil {
				result = c.doWriteRead(ctx, req, resp, opt, conn)
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

func (c *TCPClient) doWriteRead(ctx context.Context, req Request, resp Response, opt *xservice2.Option, conn net.Conn) (result error) {
	totalTimeout := opt.GetWriteTimeout() + opt.GetReadTimeout()
	ctx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()
	if result = conn.SetDeadline(time.Now().Add(totalTimeout)); result != nil {
		conn.Close()
		return result
	}
	// 暂时不将读写超时分开控制
	result = req.WriteTo(ctx, conn, opt)
	if result != nil {
		conn.Close()
		return result
	}
	_, result = resp.ReadFrom(conn)
	return result
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
