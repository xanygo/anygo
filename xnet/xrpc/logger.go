//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-09

package xrpc

import (
	"context"
	"net/http"
	"time"

	"github.com/xanygo/anygo/ds/xmetric"
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
		BeforeInvoke:   l.beforeInvoke,
		AfterWriteRead: l.afterWriteRead,
		AfterInvoke:    l.afterInvoke,
	}
}

func (l *Logger) getLogger() xlog.Logger {
	if l.Logger == nil {
		return xlog.ClientLogger()
	}
	return l.Logger
}

func (l *Logger) beforeInvoke(ctx context.Context, service string, req Request, resp Response, span xmetric.Span,
	opts ...Option) (context.Context, Request, Response, []Option) {
	ctx = xlog.NewContext(ctx)
	return ctx, req, resp, opts
}

func (l *Logger) afterWriteRead(ctx context.Context, _ string, conn *xnet.ConnNode, _ Request, resp Response, span xmetric.Span, err error) {
	item := map[string]any{
		"Cost": time.Since(span.StartTime()).String(),
	}
	if err != nil {
		item["Err"] = err.Error()
	}

	item["Conn"] = map[string]any{
		"ReadBytes":  conn.ReadBytes(),
		"WriteBytes": conn.WriteBytes(),
		"ReadCost":   conn.ReadCost().String(),
		"WriteCost":  conn.WriteCost().String(),
		"Usage":      conn.UsageCount(),
	}

	if rr, ok := resp.Unwrap().(*http.Response); ok && rr != nil {
		respInfo := map[string]any{
			"Proto":   rr.Proto,
			"Status":  rr.Status,
			"BodyLen": rr.ContentLength,
			"CT":      rr.Header.Get("Content-Type"),
		}
		item["Response"] = respInfo
	}
	span.SetAttributes(xmetric.AnyAttr("detail", item))
}

func (l *Logger) afterInvoke(ctx context.Context, _ string, _ Request, resp Response, span xmetric.Span, err error) {
	errMsg := resp.ErrMsg()
	spanInfo := xlog.Any("Trace", xmetric.Dump(span))
	// callerSkip =4 : 使日志中的 "source":<"function","file"> 定位到调用 RPC 方法的业务代码位置
	lg := l.getLogger()
	lg.Output(ctx, xlog.LevelInfo, 4, errMsg, spanInfo)
	if err != nil {
		lg.Output(ctx, xlog.LevelError, 4, err.Error(), spanInfo)
	}
}
