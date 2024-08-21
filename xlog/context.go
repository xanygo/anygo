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
	ctxKeyMeta
)

func NewContext(ctx context.Context) context.Context {
	ctx = WithMetaContext(ctx)
	return context.WithValue(ctx, ctxKeyBaggage, newBaggage())
}

func WithContext(ctx context.Context) context.Context {
	if ctx.Value(ctxKeyBaggage) != nil {
		return ctx
	}
	return NewContext(ctx)
}

func ForkContext(ctx context.Context) context.Context {
	bg := findBaggage(ctx)
	if bg == nil {
		ctx = WithMetaContext(ctx)
		bg = newBaggage()
	} else {
		bg = bg.Clone()
	}
	return context.WithValue(ctx, ctxKeyBaggage, bg)
}

func findBaggage(ctx context.Context) *baggage {
	val, _ := ctx.Value(ctxKeyBaggage).(*baggage)
	return val
}

func mustFindBaggage(ctx context.Context) *baggage {
	if val := findBaggage(ctx); val != nil {
		return val
	}
	panic("should NewContext(ctx) first")
}

func NewMetaContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyMeta, newBaggage())
}

func WithMetaContext(ctx context.Context) context.Context {
	if ctx.Value(ctxKeyMeta) != nil {
		return ctx
	}
	return NewMetaContext(ctx)
}

func ForkMetaContext(ctx context.Context) context.Context {
	bg := findMetaBaggage(ctx)
	if bg == nil {
		bg = newBaggage()
	} else {
		bg = bg.Clone()
	}
	return context.WithValue(ctx, ctxKeyMeta, bg)
}

func findMetaBaggage(ctx context.Context) *baggage {
	val, _ := ctx.Value(ctxKeyMeta).(*baggage)
	return val
}

func mustFindMetaBaggage(ctx context.Context) *baggage {
	if val := findMetaBaggage(ctx); val != nil {
		return val
	}
	panic("should NewMetaContext(ctx) first")
}

func newBaggage() *baggage {
	return &baggage{
		attrs: &xmap.Ordered[string, Attr]{},
	}
}

type baggage struct {
	attrs *xmap.Ordered[string, Attr]
}

func (cb *baggage) Clone() *baggage {
	return &baggage{
		attrs: cb.attrs.Clone(),
	}
}

func (cb *baggage) Add(attrs ...Attr) {
	for _, attr := range attrs {
		cb.attrs.Set(attr.Key, attr)
	}
}

func (cb *baggage) Delete(keys ...string) {
	cb.attrs.Delete(keys...)
}

func (cb *baggage) Attrs() []Attr {
	if cb == nil {
		return nil
	}
	return cb.attrs.Values()
}

func AddAttr(ctx context.Context, attrs ...Attr) {
	mustFindBaggage(ctx).Add(attrs...)
}

func AttrsFromCtx(ctx context.Context) []Attr {
	return findBaggage(ctx).Attrs()
}

func AddMetaAttr(ctx context.Context, attrs ...Attr) {
	mustFindMetaBaggage(ctx).Add(attrs...)
}

func MetaAttrsFromCtx(ctx context.Context) []Attr {
	return findMetaBaggage(ctx).Attrs()
}
