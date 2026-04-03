//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xrpc

import (
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xmeta"
	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/dsession"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xdial"
	"github.com/xanygo/anygo/xnet/xpolicy"
	"github.com/xanygo/anygo/xnet/xservice"
)

var _ Client = (*Feilian)(nil)

// Feilian Client 的默认实现
type Feilian struct {
	Interceptor     []Interceptor
	ServiceRegistry xservice.Registry
}

func (c *Feilian) allITs(ctx context.Context) []Interceptor {
	its := slices.Clone(globalInterceptors)
	if len(c.Interceptor) > 0 {
		its = append(its, c.Interceptor...)
	}
	if items := ITFromContext(ctx); len(items) > 0 {
		its = append(its, items...)
	}
	return its
}

func (c *Feilian) getServiceName(srv any) string {
	switch sv := srv.(type) {
	case string:
		return sv
	case xservice.Service:
		return sv.Name()
	default:
		return fmt.Sprintf("invalid-%v", sv)
	}
}

func (c *Feilian) Invoke(ctx context.Context, srv any, req Request, resp Response, opts ...Option) (result error) {
	var rootSpan xmetric.Span
	ctx, rootSpan = xmetric.Start(ctx, "invoke")
	its := c.allITs(ctx)

	cfg := &config{
		opt:      xoption.NewSimple(),
		registry: c.ServiceRegistry,
	}
	ctxOpts := OptionsFromContext(ctx)
	for _, opt := range ctxOpts {
		opt.withOption(cfg)
	}

	for _, o := range opts {
		o.withOption(cfg)
	}

	if cfg.sessionInit != nil {
		ctx = dsession.ContextWithSkip(ctx, true)
	}

	serviceName := c.getServiceName(srv)

	rootSpan.SetAttributes(
		xmetric.AnyAttr("its.len", len(its)),
		xmetric.AnyAttr("service", serviceName),
		xmetric.AnyAttr("req", req.String()),
		xmetric.AnyAttr("req.API", req.APIName()),
		xmetric.AnyAttr("req.Protocol", req.Protocol()),
	)

	service, err := cfg.getService(srv)

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

	// 设置整体的超时时间
	var timeout time.Duration
	if tv, ok := xoption.Timeout(opt); ok {
		timeout = tv
	} else {
		timeout = time.Duration(attemptTotal) * xoption.TotalTimeout(opt)
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	rootSpan.SetAttemptCount(attemptTotal)

	var retryPolicy *xpolicy.Retry
	if attemptTotal > 1 {
		retryPolicy = xoption.RetryPolicy(opt)
	}

	for attempt := range attemptTotal {
		ctxTry := ctx
		if attempt > 0 {
			ctxTry = ContextWithRetryCount(ctx, attempt)
		}
		result = c.tryOnce(ctxTry, cfg, req, resp, serviceName, service, its, opt)
		if result == nil || attempt >= attemptTotal-1 || ctxTry.Err() != nil || !retryPolicy.IsRetryable(ctxTry, req, attempt, result) {
			return result
		}
		if backoff := retryPolicy.GetBackoff(attempt); backoff > 0 {
			xctx.Sleep(ctxTry, backoff)
		}
	}
	return result
}

func (c *Feilian) tryOnce(ctx context.Context, cfg *config, req Request, resp Response, serviceName string, service xservice.Service, its []Interceptor,
	opt xoption.Reader) (result error) {
	timeout := xoption.TotalTimeout(opt)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var rootSpan xmetric.Span
	ctx, rootSpan = xmetric.Start(ctx, "TryOnce")
	defer func() {
		rootSpan.RecordError(result)
		rootSpan.End()
	}()

	for _, it := range its {
		if it.BeforePickAddress != nil {
			it.BeforePickAddress(ctx, serviceName)
		}
	}

	ap := cfg.ap
	if ap == nil {
		ap = service.Balancer()
	}
	addr, err := xbalance.Pick(ctx, ap)
	for _, it := range its {
		if it.AfterPickAddress != nil {
			it.AfterPickAddress(ctx, serviceName, addr, err)
		}
	}

	if err != nil {
		return err
	}

	entry, errPool := xdial.GroupPoolGet(ctx, service.GroupPool(), *addr)

	var conn io.ReadWriteCloser
	if errPool == nil {
		// 注册调用资源回收逻辑，之后首次调用 conn.Close()，会将 entry 对象放回对象池
		conn = entry.Borrowed()
	}

	for _, it := range its {
		if it.AfterDial != nil {
			it.AfterDial(ctx, serviceName, addr, conn, errPool)
		}
	}
	if errPool != nil {
		return errPool
	}

	if cfg.sessionInit != nil && !xmeta.HasKey(conn, xmeta.KeySessionReply) {
		reply, errSS := cfg.sessionInit.StartSession(ctx, conn, opt)
		if errSS != nil {
			return errSS
		}
		xmeta.TrySet(conn, xmeta.KeySessionReply, reply)
		rootSpan.SetAttributes(xmetric.AnyAttr("StartSession", reply))
	}

	wrCtx, wrSpan := xmetric.Start(ctx, "WriteRead")
	defer func() {
		for _, it := range its {
			if it.AfterWriteRead != nil {
				it.AfterWriteRead(ctx, serviceName, conn, req, resp, wrSpan, err)
			}
		}
		wrSpan.End()
		_ = conn.Close()
	}()

	err = c.doWriteRead(wrCtx, req, resp, opt, conn)
	wrSpan.SetAttributes(
		xmetric.AnyAttr("resp.code", resp.ErrCode()),
		xmetric.AnyAttr("resp.msg", resp.ErrMsg()),
	)
	wrSpan.RecordError(err)
	return err
}

func (c *Feilian) doWriteRead(ctx context.Context, req Request, resp Response, opt xoption.Reader, rw io.ReadWriteCloser) (err error) {
	start := time.Now()
	// 暂时不将读写超时分开控制
	err = req.WriteTo(ctx, rw, opt)
	if err != nil {
		return err
	}
	err = resp.LoadFrom(ctx, req, rw, opt)
	if err != nil {
		return fmt.Errorf("read Response %w, cost=%s", err, time.Since(start).String())
	}
	return err
}

type Interceptor struct {
	BeforeInvoke func(ctx context.Context, service string, req Request, resp Response, span xmetric.Span,
		opts ...Option) (context.Context, Request, Response, []Option)

	BeforePickAddress func(ctx context.Context, service string)
	AfterPickAddress  func(ctx context.Context, service string, node *xnet.AddrNode, err error)

	// AfterDial 拨号完成后执行，最多执行 ( retry + 1 ) * ( connectRetry +1) 次
	AfterDial func(ctx context.Context, service string, addr *xnet.AddrNode, conn io.ReadWriteCloser, err error)

	// AfterWriteRead 每 Write + Read 完成后都会执行一次，最多执行 retry+1 次
	AfterWriteRead func(ctx context.Context, service string, conn io.ReadWriteCloser, req Request, resp Response, span xmetric.Span, err error)

	// AfterInvoke 在 Invoke 执行完成后，执行一次
	AfterInvoke func(ctx context.Context, service string, req Request, resp Response, span xmetric.Span, err error)
}

var defaultClient = &xsync.OnceInit[Client]{
	New: func() Client {
		return &Feilian{}
	},
}

func DefaultClient() Client {
	return defaultClient.Load()
}

func SetDefaultClient(c Client) {
	defaultClient.Store(c)
}

func Invoke(ctx context.Context, service any, req Request, resp Response, opts ...Option) (result error) {
	return DefaultClient().Invoke(ctx, service, req, resp, opts...)
}

var globalInterceptors []Interceptor

func RegisterIT(its ...Interceptor) {
	globalInterceptors = append(globalInterceptors, its...)
}
