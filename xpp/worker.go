//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xpp

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
)

type (
	named interface {
		Name() string
	}

	starter1 interface {
		Start(ctx context.Context, cycle time.Duration) error
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
func TryStartWorker(ctx context.Context, cycle time.Duration, workers ...any) error {
	if len(workers) == 0 {
		return nil
	}
	var errs []error
	var successList []any
	for _, worker := range workers {
		if sw, ok := worker.(starter1); ok {
			if err := sw.Start(ctx, cycle); err != nil {
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
	var wg xsync.WaitGo
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

// CycleWorker 会周期性执行逻辑的 Worker,一旦启动后，后台携程会定期执行相关任务逻辑
type CycleWorker interface {
	Name() string
	Start(ctx context.Context, cycle time.Duration) error
	Stop(ctx context.Context) error
}

var _ CycleWorker = (*CycleWorkerTpl)(nil)

// CycleWorkerTpl CycleWorker 的实现
type CycleWorkerTpl struct {
	Do         func(ctx context.Context) error
	WorkerName string
	running    atomic.Bool
	tm         *time.Timer
}

func (c *CycleWorkerTpl) Name() string {
	return c.WorkerName
}

func (c *CycleWorkerTpl) Start(ctx context.Context, cycle time.Duration) error {
	if !c.running.CompareAndSwap(false, true) {
		// 已经在运行中
		return nil
	}
	if c.tm == nil {
		c.tm = time.NewTimer(cycle)
	} else {
		c.tm.Reset(cycle)
	}

	err := c.executeDo(ctx)
	if err != nil {
		c.running.Store(false)
		return err
	}
	go c.executeBG(ctx, cycle)
	return nil
}

func (c *CycleWorkerTpl) executeDo(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	return c.Do(ctx)
}

func (c *CycleWorkerTpl) executeBG(ctx context.Context, cycle time.Duration) {
	defer func() {
		c.tm.Stop()
		c.running.Store(false)
	}()
	for c.running.Load() {
		select {
		case <-ctx.Done():
			return
		case <-c.tm.C:
			c.running.Store(true)
			c.executeDo(ctx)
			c.tm.Reset(cycle)
		}
	}
}

func (c *CycleWorkerTpl) Stop(ctx context.Context) error {
	if !c.running.CompareAndSwap(true, false) {
		return nil
	}
	c.tm.Stop()
	return nil
}
