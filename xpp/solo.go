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

// SingletonWorker  有生命期限的后台任务(life 参数控制)，当达到生命期限后，启动的后台携程退出。
//
// 后台任务携程可通过 RunContext 或者 Run 触发，后台只会保留一个运行中的携程
type SingletonWorker struct {
	running atomic.Bool
}

// RunContext 尝试启动后台携程
//
// life: 后台携程最大运行时长，如 5 分钟
// cycle: do方法运行时间间隔，如 5 秒钟
func (st *SingletonWorker) RunContext(ctx context.Context, do func(ctx context.Context), life time.Duration, cycle time.Duration) {
	if !st.running.CompareAndSwap(false, true) {
		return
	}

	go st.runCtx(ctx, do, life, cycle)
}

func (st *SingletonWorker) runCtx(ctx context.Context, do func(ctx context.Context), life time.Duration, cycle time.Duration) {
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
func (st *SingletonWorker) Run(do func(), life time.Duration, cycle time.Duration) {
	if !st.running.CompareAndSwap(false, true) {
		return
	}
	go st.run(do, life, cycle)
}

func (st *SingletonWorker) run(do func(), life time.Duration, cycle time.Duration) {
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

func (st *SingletonWorker) Running() bool {
	return st.running.Load()
}

// OnDemandWorker 由业务逻辑触发，保持在后台运行的任务
type OnDemandWorker struct {
	Do    func()        // 必填
	Life  time.Duration // 每次在后台持续运行的时长,可选，默认值为 1 分钟
	Cycle time.Duration // Do 方法的运行间隔，可选，默认值为 5 秒

	running atomic.Bool
}

func (st *OnDemandWorker) Start() {
	if !st.running.CompareAndSwap(false, true) {
		return
	}
	go st.run()
}

func (st *OnDemandWorker) getLife() time.Duration {
	if st.Life > 0 {
		return st.Life
	}
	return time.Minute
}

func (st *OnDemandWorker) run() {
	tm := time.NewTimer(0)
	ctx, cancel := context.WithTimeout(context.Background(), st.getLife())
	defer func() {
		tm.Stop()
		cancel()
		st.running.Store(false)
	}()
	cycle := st.Cycle
	if cycle <= 0 {
		cycle = 5 * time.Second
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			safely.RunVoid(st.Do)
			tm.Reset(st.Cycle)
		}
	}
}

func (st *OnDemandWorker) Running() bool {
	return st.running.Load()
}
