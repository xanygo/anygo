//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-28

package xnet

import (
	"context"

	"github.com/xanygo/anygo/ds/xctx"
)

// Interceptor 拦截器接口定义，具体的实现包括 ConnInterceptor、DialerInterceptor、ResolverInterceptor
type Interceptor interface {
	Interceptor()
}

type ctxKey uint8

const (
	ctxKeyInterceptor ctxKey = iota
	ctxKeyAddr
)

// ContextWithITs 让 ctx 注册携带 Interceptor，允许注册多次。
// 最终读取的时候可以遍历向上读取所有 ctx 里注册的 Interceptor
func ContextWithITs(ctx context.Context, its ...Interceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	return xctx.WithValues(ctx, ctxKeyInterceptor, its...)
}

// ITsFromContext 从 ctx 里读取所有的 interceptor
func ITsFromContext[T Interceptor](ctx context.Context) []T {
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

// WithInterceptor  注册全局的 Interceptor
//
// 这部分 Interceptor 会在通过 ctx 注册的 Interceptor 之前执行
func WithInterceptor(its ...Interceptor) {
	for _, it := range its {
		switch vit := it.(type) {
		case *ResolverInterceptor:
			registerResolverIT(vit)
		case *DialerInterceptor:
			registerDialerITs(vit)
		case *ConnInterceptor:
			registerConnInterceptor(vit)
		default:
			panic("unsupported interceptor")
		}
	}
}
