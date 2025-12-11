//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package xsync

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/xanygo/anygo/safely"
)

// WaitGroup 所有方法都会安全的运行，会自动捕捉 panic 作为 error 返回
type WaitGroup struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error
}

func (w *WaitGroup) Go(f func()) {
	w.wg.Go(func() {
		err := safely.Run(f)
		if err == nil {
			return
		}
		w.mu.Lock()
		w.errs = append(w.errs, err)
		w.mu.Unlock()
	})
}

func (w *WaitGroup) GoCtx(ctx context.Context, f func(ctx context.Context)) {
	w.wg.Go(func() {
		err := safely.RunCtx(ctx, f)
		if err == nil {
			return
		}
		w.mu.Lock()
		w.errs = append(w.errs, err)
		w.mu.Unlock()
	})
}

func (w *WaitGroup) GoErr(f func() error) {
	w.wg.Go(func() {
		err := safely.Run(f)
		if err == nil {
			return
		}
		w.mu.Lock()
		w.errs = append(w.errs, err)
		w.mu.Unlock()
	})
}

func (w *WaitGroup) GoCtxErr(ctx context.Context, f func(ctx context.Context) error) {
	w.wg.Go(func() {
		err := safely.RunCtx(ctx, f)
		if err == nil {
			return
		}
		w.mu.Lock()
		w.errs = append(w.errs, err)
		w.mu.Unlock()
	})
}

// Wait 等待所有方法执行完成并返回所有的错误
func (w *WaitGroup) Wait() error {
	w.wg.Wait()
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.errs) == 0 {
		return nil
	}
	return errors.Join(w.errs...)
}

// WaitFirst 异步执行方法，并使用 Wait 方法获取第一个执行的错误状态
//
// 所有方法都会安全的运行，会自动捕捉 panic 作为 error 返回
type WaitFirst struct {
	sem     chan error
	once    sync.Once
	done    chan struct{}
	fnExist atomic.Bool
}

func (w *WaitFirst) init() {
	w.once.Do(func() {
		w.sem = make(chan error, 1)
		w.done = make(chan struct{})
	})
}

func (w *WaitFirst) Go(f func()) {
	w.init()
	select {
	case <-w.done:
		return
	default:
	}
	w.fnExist.Store(true)
	go func() {
		err := safely.Run(f)
		w.fire(err)
	}()
}

func (w *WaitFirst) fire(err error) {
	select {
	case w.sem <- err:
		close(w.done)
	default:
	}
}

// GoCtx 异步运行。若已有方法已运行完成，再传入方法则立即返回；若已有方法在运行中，则会触发 Cancel context
func (w *WaitFirst) GoCtx(ctx context.Context, f func(ctx context.Context)) {
	w.init()
	select {
	case <-w.done:
		return
	default:
	}
	w.fnExist.Store(true)
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		<-w.done
		cancel()
	}()
	go func() {
		err := safely.RunCtx(ctx, f)
		w.fire(err)
	}()
}

func (w *WaitFirst) GoErr(f func() error) {
	w.init()
	select {
	case <-w.done:
		return
	default:
	}
	w.fnExist.Store(true)
	go func() {
		err := safely.Run(f)
		w.fire(err)
	}()
}

// GoCtxErr 异步运行。若已有方法已运行完成，再传入方法则立即返回；若已有方法在运行中，则会触发 Cancel context
func (w *WaitFirst) GoCtxErr(ctx context.Context, f func(ctx context.Context) error) {
	w.init()
	select {
	case <-w.done:
		return
	default:
	}
	w.fnExist.Store(true)
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		<-w.done
		cancel()
	}()
	go func() {
		err := safely.RunCtx(ctx, f)
		w.fire(err)
	}()
}

// Wait 等待第一个方法执行完并立即返回执行状态
func (w *WaitFirst) Wait() error {
	if !w.fnExist.Load() {
		return errors.New("no function executed")
	}
	w.init()
	select {
	case err := <-w.sem:
		return err
	case <-w.done:
		return nil
	}
}
