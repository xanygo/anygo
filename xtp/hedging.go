//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-22

package xtp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

// Hedging 通用的请求对冲函数
type Hedging[T any] struct {
	// Main 必填，主逻辑函数
	Main func(ctx context.Context) (T, error)

	// CallNext 可选，用于前一次调用返回的结果，判断是否需要立即调用下一个函数
	CallNext func(ctx context.Context, value T, err error) bool

	fns []hedgingFn[T]
}

// Add 注册一个对冲函数
//
//	delay: 延迟绝对时间，即相对于调用 Hedging.Run 开始执行 Main 方法的时间间隔，应 >= 0
//	fn: 对冲函数
func (h *Hedging[T]) Add(delay time.Duration, fn func(ctx context.Context) (T, error)) {
	info := hedgingFn[T]{
		Delay: delay,
		Fn:    fn,
	}
	h.fns = append(h.fns, info)
}

func (h *Hedging[T]) sortFns() {
	// 按照延迟时间排序，延迟时间小的排前面
	sort.Slice(h.fns, func(i, j int) bool {
		return h.fns[i].Delay < h.fns[j].Delay
	})
	firstDelay := h.fns[0].Delay
	// 将绝对时间转换为相对时间
	for i := 1; i < len(h.fns); i++ {
		h.fns[i].Delay = h.fns[i].Delay - firstDelay
	}
}

// Run 执行并得到结果
func (h *Hedging[T]) Run(ctx context.Context) (T, error) {
	if len(h.fns) == 0 {
		return h.Main(ctx)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ret := make(chan hedgingResult[T], len(h.fns)+1)
	h.sortFns()
	go h.execute(ctx, h.Main, ret)

	tm := time.NewTimer(h.fns[0].Delay)
	defer tm.Stop()

	for i := 0; i <= len(h.fns); i++ {
		var next hedgingFn[T]
		var tc <-chan time.Time
		if i < len(h.fns) {
			next = h.fns[i]
		}
		if i > 0 && next.Fn != nil {
			tm.Reset(next.Delay)
		}
		if i < len(h.fns) {
			tc = tm.C
		}
		select {
		case <-ctx.Done():
			var emp T
			return emp, context.Cause(ctx)
		case v := <-ret:
			if next.Fn != nil && h.CallNext != nil && h.CallNext(ctx, v.Value, v.Err) {
				go h.execute(ctx, next.Fn, ret)
				continue
			}
			return v.Value, v.Err
		case <-tc:
			go h.execute(ctx, next.Fn, ret)
		}
	}
	var emp T
	// 理论不应该发生
	return emp, errors.New("bug, unexpect")
}

func (h *Hedging[T]) execute(ctx context.Context, fn func(ctx context.Context) (T, error), ret chan<- hedgingResult[T]) {
	defer func() {
		if re := recover(); re != nil {
			ret <- hedgingResult[T]{
				Err: fmt.Errorf("panic: %v", re),
			}
		}
	}()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	val, err := fn(ctx)
	select {
	case <-ctx.Done():
		ret <- hedgingResult[T]{
			Err: context.Cause(ctx),
		}
	case ret <- hedgingResult[T]{Value: val, Err: err}:
	}
}

type hedgingFn[T any] struct {
	Delay time.Duration
	Fn    func(ctx context.Context) (T, error)
}

type hedgingResult[T any] struct {
	Value T
	Err   error
}
