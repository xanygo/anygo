//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"context"
	"github.com/xanygo/anygo/xmap"
)

type ctxKey uint8

const (
	ctxKeyBaggage ctxKey = iota
)

func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyBaggage, &contextBaggage{})
}

func WithContext(ctx context.Context) context.Context {
	if ctx.Value(ctxKeyBaggage) != nil {
		return ctx
	}
	return NewContext(ctx)
}

func findBaggage(ctx context.Context) *contextBaggage {
	val, ok := ctx.Value(ctxKeyBaggage).(*contextBaggage)
	if ok {
		return val
	}
	panic("should NewContext(ctx) first")
}

type contextBaggage struct {
	values xmap.Sorted[string, Attr]
}

func (cb *contextBaggage) Add(attrs ...Attr) {
	for _, attr := range attrs {
		cb.values.Set(attr.Key, attr)
	}
}

func AddAttr(ctx context.Context, attrs ...Attr) {
	findBaggage(ctx).Add(attrs...)
}
