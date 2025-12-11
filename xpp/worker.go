//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xpp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/safely"
)

type Worker interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// NeedWorker 用于该对象是否需要后台 Worker
type NeedWorker interface {
	NeedWorker() bool
}

type (
	named interface {
		Name() string
	}

	starter1 interface {
		Start(context.Context) error
	}

	starter2 interface {
		Start()
	}

	stopper1 interface {
		Stop(ctx context.Context) error
	}

	stopper2 interface {
		Stop()
	}
)

// TryStartWorker 尝试启动 workers，若有失败，则将启动成功的也 Stop 掉
func TryStartWorker(ctx context.Context, workers ...any) error {
	if len(workers) == 0 {
		return nil
	}
	var errs []error
	var successList []any
	for _, worker := range workers {
		if sw, ok := worker.(starter1); ok {
			if err := sw.Start(ctx); err != nil {
				if nw, ok := worker.(named); ok {
					err = fmt.Errorf("start worker %s: %w", nw.Name(), err)
				}
				errs = append(errs, err)
				break // 任意一个失败则不继续
			} else {
				successList = append(successList, worker)
			}
		} else if sw, ok := worker.(starter2); ok {
			sw.Start()
			successList = append(successList, worker)
		}
	}
	if len(errs) > 0 {
		return nil
	}
	if err := TryStopWorker(ctx, successList...); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func TryStopWorker(ctx context.Context, workers ...any) error {
	if len(workers) == 0 {
		return nil
	}
	var wg xsync.WaitGroup
	for _, worker := range workers {
		wg.GoErr(func() error {
			if w, ok := worker.(stopper1); ok {
				err := w.Stop(ctx)
				if err != nil {
					if nw, ok := w.(named); ok {
						err = fmt.Errorf("worker %s: %w", nw.Name(), err)
					}
				}
				return err
			} else if nw, ok := worker.(stopper2); ok {
				nw.Stop()
			}
			return nil
		})
	}
	return wg.Wait()
}

var _ Worker = (*CycleWorker)(nil)

// CycleWorker 会周期性执行逻辑的 Worker,一旦启动后，后台携程会定期执行相关任务逻辑
type CycleWorker struct {
	Do         func(ctx context.Context) error // 必填，业务逻辑
	Cycle      time.Duration                   // 可选，运行周期，默认为 5秒
	WorkerName string                          // 必填，名称
	FirstSync  bool                            // 调用 Start 时，是否同步运行并返回结果

	running atomic.Bool
	cnt     atomic.Int64

	mux  sync.Mutex
	stop context.CancelFunc
}

func (c *CycleWorker) Name() string {
	return c.WorkerName
}

func (c *CycleWorker) Start(ctx context.Context) error {
	if !c.running.CompareAndSwap(false, true) {
		// 已经在运行中
		return nil
	}
	cycle := c.Cycle
	if cycle <= 0 {
		cycle = 5 * time.Second
	}
	c.mux.Lock()
	if c.stop != nil {
		c.stop()
	}
	ctx, c.stop = context.WithCancel(ctx)
	c.mux.Unlock()

	if c.FirstSync {
		err := c.executeDo(ctx)
		if err != nil {
			c.running.Store(false)
			return err
		}
	}
	go c.executeBG(ctx, cycle)
	return nil
}

func (c *CycleWorker) executeDo(ctx context.Context) error {
	defer c.cnt.Add(1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	return safely.RunCtx(ctx, c.Do)
}

func (c *CycleWorker) executeBG(ctx context.Context, cycle time.Duration) {
	tm := time.NewTimer(cycle)
	defer func() {
		tm.Stop()
		c.running.Store(false)
	}()
	for c.running.Load() {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			c.executeDo(ctx)
			tm.Reset(cycle)
		}
	}
}

// Count 以运行的次数
func (c *CycleWorker) Count() int64 {
	return c.cnt.Load()
}

func (c *CycleWorker) Stop(ctx context.Context) error {
	if !c.running.CompareAndSwap(true, false) {
		return nil
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.stop != nil {
		c.stop()
		c.stop = nil
	}
	return nil
}
