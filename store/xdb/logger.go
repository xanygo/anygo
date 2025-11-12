//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package xdb

import (
	"context"
	"log"

	"github.com/xanygo/anygo/xlog"
)

type Logger struct {
	Logger xlog.Logger
}

func (l *Logger) ToInterceptor() *Interceptor {
	return &Interceptor{
		After: l.after,
	}
}

func (l *Logger) getLogger() xlog.Logger {
	if l.Logger != nil {
		return l.Logger
	}
	return xlog.ClientLogger()
}

func (l *Logger) after(ctx context.Context, e Event) {
	logger := l.getLogger()
	if xlog.IsNop(logger) {
		return
	}
	attrs := []xlog.Attr{
		xlog.String("action", e.Action),
		xlog.String("client", e.Client),
		xlog.String("driver", e.Driver),
		xlog.Time("start", e.Start),
		xlog.DurationMS("cost", e.End.Sub(e.Start)),
		xlog.String("query", e.Query),
		xlog.Int("args.len", len(e.Args)),
		xlog.ErrorAttr("error", e.Error),
	}
	if e.TxID != "" {
		attrs = append(attrs, xlog.String("txID", e.TxID))
	}
	if e.StmtID != "" {
		attrs = append(attrs, xlog.String("StmtID", e.StmtID))
	}
	log.Println("after call")
	logger.Output(ctx, xlog.LevelInfo, 3, e.Action, attrs...)
}
