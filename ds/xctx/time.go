//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

package xctx

import (
	"context"
	"time"
)

// Sleep 等待 context 超时
//
// 返回值：正常完成 Sleep，返回 true；若 ctx 被提前取消，返回 false
func Sleep(ctx context.Context, dur time.Duration) bool {
	if dur <= 0 {
		return false
	}
	tm := time.NewTimer(dur)
	defer tm.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-tm.C:
		return true
	}
}
