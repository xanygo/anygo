//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xsession

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/xanygo/anygo/xctx"
	"io"
	"unsafe"
)

func NewID() string {
	bf := make([]byte, 24)
	_, _ = io.ReadFull(rand.Reader, bf)
	out := make([]byte, base64.RawURLEncoding.EncodedLen(len(bf)))
	base64.RawURLEncoding.Encode(out, bf)
	return unsafe.String(&out[0], len(out))
}

func init() {
	const idLen = 32
	id := NewID()
	if len(id) != idLen {
		panic(fmt.Errorf("sessionID=%q, expect len %d", id, idLen))
	}
}

var ctxKeySessionID = xctx.NewKey()

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeySessionID, id)
}

func IDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(ctxKeySessionID).(string)
	return val
}
