// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package xrps

import (
	"context"
	"net"

	"github.com/xanygo/anygo/xctx"
)

var ctxKeyConn = xctx.NewKey()

func ContextWithConn(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, ctxKeyConn, conn)
}

func ConnFromContext(ctx context.Context) net.Conn {
	val := ctx.Value(ctxKeyConn)
	if val == nil {
		return nil
	}
	return val.(net.Conn)
}
