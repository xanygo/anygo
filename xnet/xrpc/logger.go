//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-09

package xrpc

import (
	"context"
	"net/http"

	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet"
)

type Logger struct {
	// Logger 可选，日志输出的目标
	// 当为空时，默认使用  xlog.ClientLogger()
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

func (l *Logger) getLogger() xlog.Logger {
	if l.Logger == nil {
		return xlog.ClientLogger()
	}
	return l.Logger
}

func (l *Logger) beforeInvoke(ctx context.Context, service string, req Request, resp Response,
	opts ...Option) (context.Context, Request, Response, []Option) {
	ctx = xlog.NewContext(ctx)
	xlog.AddAttr(ctx,
		xlog.String("Service", service),
		xlog.String("Protocol", req.Protocol()),
		xlog.String("API", req.APIName()),
		xlog.String("Request", req.String()),
	)
	return ctx, req, resp, opts
}

func (l *Logger) afterPickAddress(ctx context.Context, service string, try Action, node *xnet.AddrNode, err error) {
	info := map[string]any{
		"Service": service,
		"Try":     try.TryString(),
		"Cost":    try.Cost().String(),
		"Address": node.HostPort,
	}
	if err != nil {
		info["Err"] = err.Error()
	}
	const key = "Pick"
	xlog.Append(ctx, key, info)
}

func (l *Logger) afterDial(ctx context.Context, service string, try Action, conn *xnet.ConnNode, err error) {
	info := map[string]any{
		"Service": service,
		"Remote":  conn.Addr.HostPort,
		"Cost":    try.Cost().String(),
		"Try":     try.TryString(),
	}

	if err == nil {
		info["LocalAddr"] = conn.Conn.LocalAddr().String()
		info["RemoteAddr"] = conn.Conn.RemoteAddr().String()
	} else {
		info["Err"] = err.Error()
	}
	const key = "Dial"
	xlog.Append(ctx, key, info)
}

func (l *Logger) afterWriteRead(ctx context.Context, _ string, conn *xnet.ConnNode, _ Request, resp Response, try Action, err error) {
	item := map[string]any{
		"ErrCode": resp.ErrCode(),
		"ErrMsg":  resp.ErrMsg(),
		"Cost":    try.Cost().String(),
		"Try":     try.TryString(),
	}
	if err != nil {
		item["Err"] = err.Error()
	}

	if tc, ok := conn.Conn.(*xnet.TraceConn); ok {
		item["Conn"] = map[string]any{
			"ReadBytes":  tc.ReadBytes(),
			"WriteBytes": tc.WriteBytes(),
			"ReadCost":   tc.ReadCost().String(),
			"WriteCost":  tc.WriteCost().String(),
		}
	}

	if rr, ok := resp.Unwrap().(*http.Response); ok {
		respInfo := map[string]any{
			"Proto":   rr.Proto,
			"Status":  rr.Status,
			"BodyLen": rr.ContentLength,
			"CT":      rr.Header.Get("Content-Type"),
		}
		item["Response"] = respInfo
	}

	const key = "WriteRead"

	xlog.Append(ctx, key, item)
}

func (l *Logger) afterInvoke(ctx context.Context, _ string, _ Request, resp Response, try Action, err error) {
	errMsg := resp.ErrMsg()
	attrs := []xlog.Attr{
		xlog.String("Try", try.TryString()),
		xlog.String("Cost", try.Cost().String()),
	}
	// callerSkip =3 : 使日志中的 "source":<"function","file"> 定位到调用 RPC 方法的业务代码位置
	lg := l.getLogger()
	lg.Output(ctx, xlog.LevelInfo, 3, errMsg, attrs...)
	if err != nil {
		lg.Output(ctx, xlog.LevelError, 3, err.Error(), attrs...)
	}
}
