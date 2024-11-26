//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xsession

import (
	"context"
	"time"

	"github.com/xanygo/anygo/xcodec/xbase62"
	"github.com/xanygo/anygo/xctx"
	"github.com/xanygo/anygo/xstr"
)

func NewID() string {
	tm := xbase62.EncodeInt64(time.Now().Unix() - 1730000000)
	id := xstr.RandomN(8)
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
