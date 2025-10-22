//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-08

package xbalance

import (
	"context"

	"github.com/xanygo/anygo/ds/xctx"
)

var ctxKeyTarget = xctx.NewKey()

func ContextWithReader(ctx context.Context, ap Reader) context.Context {
	return context.WithValue(ctx, ctxKeyTarget, ap)
}

func ReaderFromContext(ctx context.Context) Reader {
	val, _ := ctx.Value(ctxKeyTarget).(Reader)
	return val
}
