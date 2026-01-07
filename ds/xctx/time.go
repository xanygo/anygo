//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

package xctx

import (
	"context"
	"time"
)

// Sleep 等待 context 超时
func Sleep(ctx context.Context, dur time.Duration) {
	if dur <= 0 {
		return
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, dur)
	defer cancel()
	<-ctx.Done()
}
