//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely

import (
	"context"
	"testing"

	"github.com/fsgo/fst"
)

func TestRunVoid(t *testing.T) {
	var num int
	RunVoid(func() {
		num++
		panic("hello")
	})
	var reNum int
	OnRecovered(func(ctx context.Context, re *PanicErr) {
		reNum++
	})
	RunVoid(func() error {
		num++
		panic("hello")
	})
	fst.Equal(t, 2, num)
	fst.Equal(t, 1, reNum)
}

func TestRun(t *testing.T) {
	var num int
	fst.NoError(t, Run(func() { num++ }))
	fst.NoError(t, Run(func() error {
		num++
		return nil
	}))
	fst.Equal(t, 2, num)

	fst.Error(t, Run(func() {
		num++
		panic("hello")
	}))
	fst.Error(t, Run(func() error {
		num++
		panic("hello")
	}))
	fst.Equal(t, 4, num)
}

func TestRunCtxVoid(t *testing.T) {
	var num int
	RunCtxVoid(context.Background(), func(ctx context.Context) {
		num++
		panic("hello")
	})
	RunCtxVoid(context.Background(), func(ctx context.Context) error {
		num++
		panic("hello")
	})
	fst.Equal(t, 2, num)
}

func TestRunCtx(t *testing.T) {
	var num int
	fst.NoError(t, RunCtx(context.Background(), func(ctx context.Context) { num++ }))
	fst.NoError(t, RunCtx(context.Background(), func(ctx context.Context) error {
		num++
		return nil
	}))
	fst.Equal(t, 2, num)

	fst.Error(t, RunCtx(context.Background(), func(ctx context.Context) {
		num++
		panic("hello")
	}))
	fst.Error(t, RunCtx(context.Background(), func(ctx context.Context) error {
		num++
		panic("hello")
	}))
	fst.Equal(t, 4, num)
}
