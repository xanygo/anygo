//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xsession

import (
	"context"
	"hash/crc32"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/xcodec/xbase"
)

// NewID 生成一个新的 SessionID
func NewID() string {
	tm := xbase.Base62.EncodeInt64(time.Now().Unix() - 1730000000)
	id := xstr.RandNChar(8)
	str := tm + "|" + id
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	hi := crc32.ChecksumIEEE(bf)
	hs := xbase.Base62.EncodeInt64(int64(hi))
	return str + "|" + hs
}

const idMinLen = 12

func IsValidID(id string) bool {
	if len(id) < idMinLen {
		return false
	}
	head, sign, found := xstr.CutLastN(id, "|", 0)
	if !found {
		return false
	}
	bf := unsafe.Slice(unsafe.StringData(head), len(head))
	hi := crc32.ChecksumIEEE(bf)
	hs := xbase.Base62.EncodeInt64(int64(hi))
	return sign == hs
}

var ctxKeySessionID = xctx.NewKey()

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeySessionID, id)
}

func IDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(ctxKeySessionID).(string)
	return val
}
