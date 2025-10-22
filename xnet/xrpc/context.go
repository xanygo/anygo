//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xrpc

import (
	"context"

	xctx2 "github.com/xanygo/anygo/ds/xctx"
)

var (
	ctxTCPITKey  = xctx2.NewKey()
	ctxOptionKey = xctx2.NewKey()
)

func ContextWithTCPIT(ctx context.Context, its ...TCPInterceptor) context.Context {
	return xctx2.WithValues(ctx, ctxTCPITKey, its...)
}

func TCPITFromContext(ctx context.Context) []TCPInterceptor {
	return xctx2.Values[*xctx2.Key, TCPInterceptor](ctx, ctxTCPITKey, true)
}

// ContextWithOption 将 Options 临时存储到 context 中去。
// 支持调用多次，最终使用 OptionsFromContext 或读取到所有的 options。
//
// 如:
//
//		ctx=ContextWithOption(ctx,OptRetry(1)) // 第 1 次
//		ctx=ContextWithOption(ctx,OptReadTimeout(time.Second)) // 第 2 次
//
//	 最终使用 OptionsFromContext(ctx) 会同时读取到 OptRetry(1) 和 OptReadTimeout(time.Second)
func ContextWithOption(ctx context.Context, opts ...Option) context.Context {
	if len(opts) == 0 {
		return ctx
	}
	return xctx2.WithValues(ctx, ctxOptionKey, opts...)
}

// OptionsFromContext 读取 ContextWithOption 设置携带在 ctx 中的 Option。
// Client.Invoke 方法的实现默认已调用
func OptionsFromContext(ctx context.Context) []Option {
	return xctx2.Values[*xctx2.Key, Option](ctx, ctxOptionKey, true)
}
