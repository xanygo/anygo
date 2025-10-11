//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
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
	ctx, rootSpan := xmetric.Start(ctx, "invoke")
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

	rootSpan.SetAttributes(
		xmetric.AnyAttr("its.len", len(its)),
		xmetric.AnyAttr("service", serviceName),
		xmetric.AnyAttr("req", req.String()),
		xmetric.AnyAttr("req.API", req.APIName()),
		xmetric.AnyAttr("req.Protocol", req.Protocol()),
	)

	service, err := cfg.getService(srv, c.ServiceRegistry)

	defer func() {
		rootSpan.RecordError(result)
		rootSpan.End()
		for _, it := range its {
			if it.AfterInvoke != nil {
				it.AfterInvoke(ctx, serviceName, req, resp, rootSpan, result)
			}
		}
	}()

	for _, it := range its {
		if it.BeforeInvoke == nil {
			continue
		}
		ctx, req, resp, opts = it.BeforeInvoke(ctx, serviceName, req, resp, rootSpan, opts...)
	}

	if err != nil {
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

	attemptTotal := xoption.Retry(opt) + 1

	rootSpan.SetAttemptCount(attemptTotal)

	for attemptNo := 0; attemptNo < attemptTotal; attemptNo++ {
		result = c.tryOnce(ctx, req, resp, serviceName, service, its, opt)
		if result == nil {
			break
		}
		if err1 := ctx.Err(); err1 != nil {
			break
		}
	}
	return result
}

func (c *TCP) tryOnce(ctx context.Context, req Request, resp Response, serviceName string, service xservice.Service, its []TCPInterceptor,
	opt xoption.Reader) (result error) {
	ctx, rootSpan := xmetric.Start(ctx, "TryOnce")
	defer func() {
		rootSpan.RecordError(result)
		rootSpan.End()
	}()

	for _, it := range its {
		if it.BeforePickAddress != nil {
			it.BeforePickAddress(ctx, serviceName)
		}
	}
	addr, err := xbalance.Pick(ctx, service.Balancer())
	for _, it := range its {
		if it.AfterPickAddress != nil {
			it.AfterPickAddress(ctx, serviceName, addr, err)
		}
	}

	if err != nil {
		return err
	}

	entry, err := xdial.GroupPoolGet(ctx, service.Pool(), *addr)

	var conn *xnet.ConnNode
	if err == nil {
		conn = entry.Object()
		var once sync.Once
		conn.OnClose = func() error {
			once.Do(func() {
				conn.OnClose = nil
				entry.Release(result)
			})
			return nil
		}
	} else if conn == nil {
		conn = &xnet.ConnNode{
			Addr: *addr,
		}
	}

	for _, it := range its {
		if it.AfterDial != nil {
			it.AfterDial(ctx, serviceName, conn, err)
		}
	}
	if err != nil {
		return err
	}

	wrCtx, wrSpan := xmetric.Start(ctx, "WriteRead")
	defer wrSpan.End()

	err = c.doWriteRead(wrCtx, req, resp, opt, conn)
	wrSpan.SetAttributes(
		xmetric.AnyAttr("resp.code", resp.ErrCode()),
		xmetric.AnyAttr("resp.msg", resp.ErrMsg()),
	)
	wrSpan.RecordError(err)
	for _, it := range its {
		if it.AfterWriteRead != nil {
			it.AfterWriteRead(ctx, serviceName, conn, req, resp, wrSpan, err)
		}
	}

	return err
}

func (c *TCP) doWriteRead(ctx context.Context, req Request, resp Response, opt xoption.Reader, conn *xnet.ConnNode) (err error) {
	defer conn.Close()
	totalTimeout := xoption.WriteReadTimeout(opt)
	ctx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()
	if err = conn.SetDeadline(time.Now().Add(totalTimeout)); err != nil {
		return err
	}

	start := time.Now()
	// 暂时不将读写超时分开控制
	err = req.WriteTo(ctx, conn, opt)
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
	BeforeInvoke func(ctx context.Context, service string, req Request, resp Response, span xmetric.Span,
		opts ...Option) (context.Context, Request, Response, []Option)

	BeforePickAddress func(ctx context.Context, service string)
	AfterPickAddress  func(ctx context.Context, service string, node *xnet.AddrNode, err error)

	// AfterDial 拨号完成后执行，最多执行 ( retry + 1 ) * ( connectRetry +1) 次
	AfterDial func(ctx context.Context, service string, conn *xnet.ConnNode, err error)

	// AfterWriteRead 每 Write + Read 完成后都会执行一次，最多执行 retry+1 次
	AfterWriteRead func(ctx context.Context, service string, conn *xnet.ConnNode, req Request, resp Response, span xmetric.Span, err error)

	// AfterInvoke 在 Invoke 执行完成后，执行一次
	AfterInvoke func(ctx context.Context, service string, req Request, resp Response, span xmetric.Span, err error)
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
