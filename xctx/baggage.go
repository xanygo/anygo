//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package xctx

import (
	"context"

	"github.com/xanygo/anygo/internal/zslice"
)

type baggage[K comparable, V any] struct {
	ctx    context.Context
	values []V
}

func (b *baggage[K, V]) All(key K) []V {
	var vs []V
	if pic, ok := b.ctx.Value(key).(*baggage[K, V]); ok {
		vs = pic.All(key)
	}
	if len(vs) == 0 {
		return b.values
	} else if len(b.values) == 0 {
		return vs
	}
	return zslice.Merge(vs, b.values)
}

// WithValues 让 ctx 携带多个值，支持多次调用，最终可以使用 Values 读取出最新的值或者递归读取所有值
func WithValues[K comparable, V any](ctx context.Context, key K, vs ...V) context.Context {
	if len(vs) == 0 {
		return ctx
	}
	val := &baggage[K, V]{
		ctx:    ctx,
		values: vs,
	}
	return context.WithValue(ctx, key, val)
}

// Values 读取使用 WithValues 设置的值，先设置的值会排在前面
//
// recursion: 是否递归读取所有的值。即是否在 ctx 中，向上查找相同 key 设置的值
func Values[K comparable, V any](ctx context.Context, key K, recursion bool) []V {
	if bg, ok := ctx.Value(key).(*baggage[K, V]); ok {
		if recursion {
			return bg.All(key)
		}
		return bg.values
	}
	return nil
}
