//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely

import (
	"context"
	"runtime"
	"runtime/debug"
)

type FnType1 interface {
	func() | func() error
}

// RunVoid 执行传入的方法，会自动 recover panic
func RunVoid[T FnType1](fn T) {
	_ = Run(fn)
}

// Run 执行传入的方法，会自动 recover panic，若 panic 则 panic 信息会以 error 返回
func Run[T FnType1](fn T) (err error) {
	defer func() {
		if re := recover(); re != nil {
			_, file, line, _ := runtime.Caller(2)
			pe := &PanicErr{
				ID:    NewRecoverID(),
				Panic: re,
				Stack: debug.Stack(),
				File:  file,
				Line:  line,
			}
			err = pe
			RecoveredPE(pe)
		}
	}()
	var obj any = fn
	switch val := obj.(type) {
	case func():
		val()
	case func() error:
		err = val()
	}
	return err
}

// WrapVoid 包装传入的方法，使其运行过程中的 panic 会被自动 recover
func WrapVoid[T FnType1](fn T) func() {
	return func() {
		RunVoid(fn)
	}
}

// Wrap 包装传入的方法，使其运行过程中的 panic 会被自动 recover
func Wrap[T FnType1](fn T) func() error {
	return func() error {
		return Run(fn)
	}
}

type FnType2 interface {
	func(ctx context.Context) | func(ctx context.Context) error
}

// RunCtxVoid 执行传入的方法，会自动 recover panic
func RunCtxVoid[T FnType2](ctx context.Context, fn T) {
	_ = RunCtx(ctx, fn)
}

// RunCtx 执行传入的方法，会自动 recover panic，若 panic 则 panic 信息会以 error 返回
func RunCtx[T FnType2](ctx context.Context, fn T) (err error) {
	defer func() {
		if re := recover(); re != nil {
			_, file, line, _ := runtime.Caller(2)
			pe := &PanicErr{
				ID:    NewRecoverID(),
				Panic: re,
				Stack: debug.Stack(),
				File:  file,
				Line:  line,
			}
			err = pe
			RecoveredPECtx(ctx, pe)
		}
	}()
	var obj any = fn
	switch val := obj.(type) {
	case func(ctx2 context.Context):
		val(ctx)
	case func(ctx2 context.Context) error:
		err = val(ctx)
	}
	return err
}

// WrapCtxVoid 包装传入的方法，使其运行过程中的 panic 会被自动 recover
func WrapCtxVoid[T FnType2](fn T) func(ctx context.Context) {
	return func(ctx context.Context) {
		RunCtxVoid(ctx, fn)
	}
}
