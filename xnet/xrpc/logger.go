//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-09

package xrpc

import (
	"context"
	"net"

	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet/xnaming"
)

type Logger struct {
	Logger xlog.Logger
}

func (l *Logger) Interceptor() TCPInterceptor {
	return TCPInterceptor{
		BeforeInvoke:     l.beforeInvoke,
		AfterPickAddress: l.afterPickAddress,
		AfterWriteRead:   l.afterWriteRead,
		AfterInvoke:      l.afterInvoke,
		AfterDial:        l.afterDial,
	}
}

var protocolHTTP = xlog.String("protocol", "HTTP")

func (l *Logger) beforeInvoke(ctx context.Context, service string, req Request, resp Response,
	opts ...Option) (context.Context, Request, Response, []Option, error) {
	ctx = xlog.NewContext(ctx)
	xlog.AddAttr(ctx,
		xlog.String("service", service),
		protocolHTTP,
		xlog.String("api", req.APIName()),
	)
	return ctx, req, resp, opts, nil
}

func (l *Logger) afterPickAddress(ctx context.Context, _ string, try TryInfo, node xnaming.Node, err error) {
	if node != nil {
		xlog.AddAttr(ctx, xlog.String("remote", node.Addr().String()))
	} else if err != nil {
		xlog.AddAttr(ctx, xlog.ErrorAttr("pickErr", err))
	}
}

func (l *Logger) afterDial(ctx context.Context, service string, node xnaming.Node, try TryInfo, conn net.Conn, err error) {
	current := map[string]any{
		"Remote": node.Addr().String(),
		"Cost":   try.Cost().Milliseconds(),
		"Try":    try.String(),
	}

	if err == nil {
		current["Local"] = conn.LocalAddr().String()
	} else {
		current["Err"] = err.Error()
	}

	var items []any
	const key = "dial"
	attr, ok := xlog.FindAttrFromCtx(ctx, key)
	if ok {
		items = attr.Value.Any().([]any)
	}
	items = append(items, current)
	attr = xlog.Any(key, items)
	xlog.AddAttr(ctx, attr)
}

func (l *Logger) afterWriteRead(ctx context.Context, _ string, _ Request, resp Response, try TryInfo, err error) {
	item := map[string]any{
		"ErrCode": resp.ErrCode(),
		"ErrMsg":  resp.ErrMsg(),
		"Cost":    try.Cost().Milliseconds(),
		"Try":     try.String(),
	}
	if err != nil {
		item["Err"] = err.Error()
	}

	const key = "writeRead"

	var items []any

	attr, ok := xlog.FindAttrFromCtx(ctx, key)
	if ok {
		items = attr.Value.Any().([]any)
	}
	items = append(items, item)
	attr1 := xlog.Any(key, items)
	xlog.AddAttr(ctx, attr1)
}

func (l *Logger) afterInvoke(ctx context.Context, _ string, _ Request, resp Response, try TryInfo, err error) {
	errMsg := resp.ErrMsg()
	attrs := []xlog.Attr{
		xlog.String("try", try.String()),
		xlog.Int64("cost", try.Cost().Milliseconds()),
	}
	// callerSkip =3 : 使日志中的 "source":<"function","file"> 定位到调用 RPC 方法的业务代码位置
	l.Logger.Output(ctx, xlog.LevelInfo, 3, errMsg, attrs...)
	if err != nil {
		l.Logger.Output(ctx, xlog.LevelError, 3, err.Error(), attrs...)
	}
}
