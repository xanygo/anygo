//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xsession

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/internal/znum"
	"github.com/xanygo/anygo/xctx"
)

func NewID() string {
	bf := make([]byte, 24)
	_, _ = io.ReadFull(rand.Reader, bf)
	out := make([]byte, base64.RawURLEncoding.EncodedLen(len(bf)))
	base64.RawURLEncoding.Encode(out, bf)
	tm := znum.FormatIntB62(time.Now().Unix() - 1730000000)
	return tm + "|" + unsafe.String(&out[0], len(out))
}

var ctxKeySessionID = xctx.NewKey()

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeySessionID, id)
}

func IDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(ctxKeySessionID).(string)
	return val
}
