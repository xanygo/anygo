// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package xrps

import (
	"context"

	"github.com/xanygo/anygo/ds/xctx"
)

var ctxKeyRW = xctx.NewKey()

func ContextWithConn[C any](ctx context.Context, conn C) context.Context {
	return context.WithValue(ctx, ctxKeyRW, conn)
}

func ConnFromContext[C any](ctx context.Context) C {
	val, _ := ctx.Value(ctxKeyRW).(C)
	return val
}
