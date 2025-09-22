//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xsession

import (
	"context"
	"time"

	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/xcodec/xbase"
	"github.com/xanygo/anygo/xctx"
)

// NewID 生成一个新的 SessionID
func NewID() string {
	tm := xbase.Base62.EncodeInt64(time.Now().Unix() - 1730000000)
	id := xstr.RandNChar(8)
	return tm + "|" + id
}

var ctxKeySessionID = xctx.NewKey()

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeySessionID, id)
}

func IDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(ctxKeySessionID).(string)
	return val
}
