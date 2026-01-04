//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpp

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/safely"
)

// SoloTask  有生命期限的后台任务(life 参数控制)，当达到生命期限后，启动的后台携程退出。
//
// 后台任务携程可通过 RunContext 或者 Run 触发，后台只会保留一个运行中的携程
type SoloTask struct {
	running atomic.Bool
}

// RunContext 尝试启动后台携程
//
// life: 后台携程最大运行时长，如 5 分钟
// cycle: do方法运行时间间隔，如 5 秒钟
func (st *SoloTask) RunContext(ctx context.Context, do func(ctx context.Context), life time.Duration, cycle time.Duration) {
	if !st.running.CompareAndSwap(false, true) {
		return
	}

	go st.runCtx(ctx, do, life, cycle)
}

func (st *SoloTask) runCtx(ctx context.Context, do func(ctx context.Context), life time.Duration, cycle time.Duration) {
	ctx1, cancel := context.WithTimeout(ctx, life)
	tm := time.NewTimer(0)
	defer func() {
		tm.Stop()
		cancel()
		st.running.Store(false)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			safely.RunCtx(ctx1, do)
			tm.Reset(cycle)
		}
	}
}

// Run 尝试启动后台携程
//
// life: 后台携程最大运行时长，如 5 分钟
// cycle: do方法运行时间间隔，如 5 秒钟
func (st *SoloTask) Run(do func(), life time.Duration, cycle time.Duration) {
	if !st.running.CompareAndSwap(false, true) {
		return
	}
	go st.run(do, life, cycle)
}

func (st *SoloTask) run(do func(), life time.Duration, cycle time.Duration) {
	tm := time.NewTimer(0)
	ctx, cancel := context.WithTimeout(context.Background(), life)
	defer func() {
		tm.Stop()
		cancel()
		st.running.Store(false)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			safely.RunVoid(do)
			tm.Reset(cycle)
		}
	}
}

func (st *SoloTask) Running() bool {
	return st.running.Load()
}
