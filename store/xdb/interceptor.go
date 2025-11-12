//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package xdb

import (
	"context"
	"time"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/internal/zslice"
)

type Event struct {
	Client string
	Driver string
	Action string
	Start  time.Time
	End    time.Time
	Query  string
	Args   []any
	Error  error
	TxID   string
	StmtID string
}

type Interceptor struct {
	After func(ctx context.Context, e Event)
}

var globalInterceptors interceptors

func RegisterIT(its ...*Interceptor) {
	globalInterceptors = append(globalInterceptors, its...)
}

var ctxKeyIt = xctx.NewKey()

func ContextWithIT(ctx context.Context, its ...*Interceptor) context.Context {
	return xctx.WithValues(ctx, ctxKeyIt, its...)
}

func allInterceptors(ctx context.Context) interceptors {
	its := xctx.Values[*xctx.Key, *Interceptor](ctx, ctxKeyIt, true)
	return zslice.SafeMerge(globalInterceptors, its)
}

type interceptors []*Interceptor

func (its interceptors) CallAfter(ctx context.Context, e Event) {
	for _, it := range its {
		if it != nil && it.After != nil {
			it.After(ctx, e)
		}
	}
}
