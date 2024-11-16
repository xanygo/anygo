//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-16

package xlog

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/internal/znum"
)

const fieldLogID = "logid"

var logid atomic.Int64

func NewLogID() string {
	num1 := time.Now().Unix() - 1731686400
	return znum.FormatIntB62(num1) + "-" + znum.Random1() + "-" + znum.FormatIntB62(logid.Add(1))
}

func WithLogID(ctx context.Context, logID string) {
	AddMetaAttr(ctx, String(fieldLogID, logID))
}

func FindLogID(ctx context.Context) string {
	f, ok := FindMetaAttrFromCtx(ctx, fieldLogID)
	if !ok {
		return ""
	}
	val, _ := f.Value.Any().(string)
	return val
}
