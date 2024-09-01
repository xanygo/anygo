//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-28

package xnet

import (
	"context"

	"github.com/xanygo/anygo/internal/zslice"
	"github.com/xanygo/anygo/xctx"
)

type Interceptor interface {
	Interceptor()
}

type ctxKey uint8

const (
	ctxKeyInterceptor ctxKey = iota
	ctxKeyAddr
)

// ContextWithInterceptor 让 ctx 注册携带 ConnInterceptor，允许注册多次，最终读取的时候遍历向上读取所有 ctx 里注册的
func ContextWithInterceptor(ctx context.Context, its ...Interceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	return xctx.WithValues(ctx, ctxKeyInterceptor, its...)
}

// InterceptorFromContext 从 ctx 里读取所有的 interceptor
func InterceptorFromContext[T Interceptor](ctx context.Context) []T {
	its := xctx.Values[ctxKey, Interceptor](ctx, ctxKeyInterceptor, true)
	if len(its) == 0 {
		return nil
	}
	result := make([]T, 0, len(its))
	for _, it := range its {
		if val, ok := it.(T); ok {
			result = append(result, val)
		}
	}
	return result
}

var globalIts []Interceptor

// WithInterceptor  注册全局的 Interceptor
//
// 这部分 Interceptor 会在通过 ctx 注册的 Interceptor 之前执行
func WithInterceptor(its ...Interceptor) {
	globalIts = append(globalIts, its...)
}

func InterceptorFromGlobal[T Interceptor]() []T {
	if len(globalIts) == 0 {
		return nil
	}
	its := make([]T, 0, len(globalIts))
	for _, it := range globalIts {
		if val, ok := it.(T); ok {
			its = append(its, val)
		}
	}
	return its
}

// Interceptors 读取全局拦截器、ctx 里的拦截器，并和传入的一起合并在一起并返回
//
// 上述 3 者合并的顺序依次为：全局拦截器、local 拦截器、ctx 拦截器
func Interceptors[T Interceptor](ctx context.Context, local []T) []T {
	its1 := InterceptorFromGlobal[T]()
	its3 := InterceptorFromContext[T](ctx)
	return zslice.Merge(its1, local, its3)
}
