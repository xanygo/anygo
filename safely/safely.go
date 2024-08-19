//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely

import (
	"context"
)

type FnType1 interface {
	func() | func() error
}

func RunVoid[T FnType1](fn T) {
	_ = Run(fn)
}

func Run[T FnType1](fn T) (err error) {
	defer func() {
		if re := recover(); re != nil {
			pe := &PanicErr{
				Re: re,
			}
			err = pe
			RecoveredVoid(pe)
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

func WrapVoid[T FnType1](fn T) func() {
	return func() {
		RunVoid(fn)
	}
}

// Wrap 包装 fn，使其自动 recover panic
func Wrap[T FnType1](fn T) func() error {
	return func() error {
		return Run(fn)
	}
}

type FnType2 interface {
	func(ctx context.Context) | func(ctx context.Context) error
}

func RunCtxVoid[T FnType2](ctx context.Context, fn T) {
	_ = RunCtx(ctx, fn)
}

func RunCtx[T FnType2](ctx context.Context, fn T) (err error) {
	defer func() {
		if re := recover(); re != nil {
			pe := &PanicErr{
				Re: re,
			}
			err = pe
			RecoveredCtx(ctx, pe)
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

func WrapCtxVoid[T FnType2](fn T) func(ctx context.Context) {
	return func(ctx context.Context) {
		RunCtxVoid(ctx, fn)
	}
}
