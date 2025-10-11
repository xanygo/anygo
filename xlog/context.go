//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xlog

import (
	"context"

	"github.com/xanygo/anygo/ds/xmap"
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

// IsContext 是否已初始化的 普通 context
func IsContext(ctx context.Context) bool {
	return ctx.Value(ctxKeyBaggage) != nil
}

// IsMetaContext 是否已初始化的 meta context
func IsMetaContext(ctx context.Context) bool {
	return ctx.Value(ctxKeyMeta) != nil
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
		attrs: &xmap.OrderedSync[string, Attr]{},
	}
}

type baggage struct {
	attrs *xmap.OrderedSync[string, Attr]
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
	if cb == nil || len(keys) == 0 {
		return
	}
	cb.attrs.Delete(keys...)
}

func (cb *baggage) Attrs() []Attr {
	if cb == nil {
		return nil
	}
	return cb.attrs.Values()
}

func (cb *baggage) Find(key string) (Attr, bool) {
	if cb == nil {
		return Attr{}, false
	}
	return cb.attrs.Get(key)
}

// AddAttr 让 ctx 携带一些日志字段，若字段同名( Key 相同)，则新的会覆盖旧的。
// 在使用前，ctx 应使用 NewContext 或者 WithContext 初始化，否则会 panic。
func AddAttr(ctx context.Context, attrs ...Attr) {
	mustFindBaggage(ctx).Add(attrs...)
}

// DeleteAttr 删除 ctx 携带的日志字段
func DeleteAttr(ctx context.Context, keys ...string) {
	findBaggage(ctx).Delete(keys...)
}

func AttrsFromCtx(ctx context.Context) []Attr {
	return findBaggage(ctx).Attrs()
}

func FindAttrFromCtx(ctx context.Context, key string) (Attr, bool) {
	return findBaggage(ctx).Find(key)
}

// AddMetaAttr 让 ctx 携带一些 meta 日志字段，若字段同名( Key 相同)，则新的会覆盖旧的。
// 在使用前，ctx 应使用 NewContext 或者 NewMetaContext 或者 WithMetaContext 初始化，否则会 panic。
func AddMetaAttr(ctx context.Context, attrs ...Attr) {
	mustFindMetaBaggage(ctx).Add(attrs...)
}

// DeleteMetaAttr 删除 ctx 携带的 meta 日志字段
func DeleteMetaAttr(ctx context.Context, keys ...string) {
	findMetaBaggage(ctx).Delete(keys...)
}

func MetaAttrsFromCtx(ctx context.Context) []Attr {
	return findMetaBaggage(ctx).Attrs()
}

func FindMetaAttrFromCtx(ctx context.Context, key string) (Attr, bool) {
	return findMetaBaggage(ctx).Find(key)
}

func AllAttrsFromCtx(ctx context.Context) []Attr {
	var attrs []Attr
	values := MetaAttrsFromCtx(ctx)
	attrs = append(attrs, values...)
	values = AttrsFromCtx(ctx)
	attrs = append(attrs, values...)
	return attrs
}

// Append （普通日志字段）找到日志字段名为 key 的 Attr，往起元素列表追加元素
func Append(ctx context.Context, key string, values ...any) {
	var items []any

	attr, ok := FindAttrFromCtx(ctx, key)
	if ok {
		items = attr.Value.Any().([]any)
	}
	items = append(items, values...)
	attr1 := Any(key, items)
	AddAttr(ctx, attr1)
}

// AppendMeta （元信息日志字段）找到日志字段名为 key 的 Attr，往起元素列表追加元素
func AppendMeta(ctx context.Context, key string, values ...any) {
	var items []any

	attr, ok := FindMetaAttrFromCtx(ctx, key)
	if ok {
		items = attr.Value.Any().([]any)
	}
	items = append(items, values...)
	attr1 := Any(key, items)
	AddMetaAttr(ctx, attr1)
}
