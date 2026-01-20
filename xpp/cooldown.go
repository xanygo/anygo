//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-08

package xpp

import (
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/safely"
)

// CooldownRunner 需要定期冷却的的 Runner，如每 5 分钟运行一次
type CooldownRunner struct {
	running atomic.Bool
	timer   atomic.Pointer[time.Timer]
}

// Run 尝试执行，若没有到达运行条件（冷却时间未到）则立即返回
// 若传入的 interval<=0 则使用默认值 5 minute
func (c *CooldownRunner) Run(interval time.Duration, fn func()) {
	if !c.running.CompareAndSwap(false, true) {
		return
	}
	go c.execute(interval, fn)
}

func (c *CooldownRunner) execute(interval time.Duration, fn func()) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	safely.RunVoid(fn)
	if old := c.timer.Swap(nil); old != nil {
		old.Stop()
	}
	timer := time.AfterFunc(interval, func() {
		c.running.Store(false)
	})
	c.timer.Store(timer)
}
