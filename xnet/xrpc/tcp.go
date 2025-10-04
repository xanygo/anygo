//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"slices"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ Client = (*TCP)(nil)

type TCP struct {
	Interceptor     []TCPInterceptor
	ServiceRegistry xservice.Registry
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

	tryTotal := xoption.Retry(opt) + 1

	var addr *xnet.AddrNode
	for try := 0; try < tryTotal; try++ {
		pickAction := NewAction("Pick", tryTotal)
		pickAction.TryIndex = try
		for _, it := range its {
			if it.BeforePickAddress != nil {
				it.BeforePickAddress(ctx, serviceName, pickAction)
			}
		}
		addr, result = service.Balancer().Pick(ctx)
		pickAction.End = time.Now()
		for _, it := range its {
			if it.AfterPickAddress != nil {
				it.AfterPickAddress(ctx, serviceName, pickAction, addr, result)
			}
		}
		if result != nil {
			continue
		}

		dialAction := NewAction("Dial", tryTotal)
		dialAction.TryIndex = try
		var conn *xnet.ConnNode
		conn, result = xdial.Connect(ctx, service.Connector(), *addr, opt)
		dialAction.End = time.Now()

		for _, it := range its {
			if it.AfterDial != nil {
				it.AfterDial(ctx, serviceName, dialAction, conn, err)
			}
		}

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
	var conn net.Conn = node.NetConn()
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
	maxBody := xoption.MaxResponseSize(opt)
	rd := io.LimitReader(conn, maxBody)
	err = resp.LoadFrom(ctx, req, rd, opt)
	if err != nil {
		return fmt.Errorf("%w, cost=%s", err, time.Since(start).String())
	}
	return err
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
