//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-11

package xsync_test

import (
	"context"
	"io"
	"sync/atomic"
	"testing"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xt"
)

func TestWaitFirst(t *testing.T) {
	t.Run("case 1 empty", func(t *testing.T) {
		var wg xsync.WaitFirst
		xt.Error(t, wg.Wait())
	})

	t.Run("case 2 void nil", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.Go(func() {
			cnt.Add(1)
		})
		xt.NoError(t, wg.Wait())

		xt.Equal(t, int32(1), cnt.Load())
		wg.Go(func() {
			cnt.Add(1)
		})
		xt.Equal(t, int32(1), cnt.Load())
	})

	t.Run("case 3 void panic", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.Go(func() {
			cnt.Add(1)
			panic("hello")
		})
		xt.Error(t, wg.Wait())

		xt.Equal(t, int32(1), cnt.Load())
		wg.Go(func() {
			cnt.Add(1)
			panic("hello")
		})
		xt.Equal(t, int32(1), cnt.Load())
	})

	t.Run("case 4 error nil", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.GoErr(func() error {
			cnt.Add(1)
			return nil
		})
		xt.NoError(t, wg.Wait())
		xt.Equal(t, int32(1), cnt.Load())

		wg.GoErr(func() error {
			cnt.Add(1)
			return nil
		})
		xt.Equal(t, int32(1), cnt.Load())
	})

	t.Run("case 5 error", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.GoErr(func() error {
			cnt.Add(1)
			return io.EOF
		})
		xt.Error(t, wg.Wait())
		xt.Equal(t, int32(1), cnt.Load())

		wg.GoErr(func() error {
			cnt.Add(1)
			return io.EOF
		})
		xt.Equal(t, int32(1), cnt.Load())
	})

	t.Run("case 6 ctx nil", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.GoCtx(context.Background(), func(ctx context.Context) {
			cnt.Add(1)
		})
		xt.NoError(t, wg.Wait())
		xt.Equal(t, int32(1), cnt.Load())

		wg.GoCtx(context.Background(), func(ctx context.Context) {
			cnt.Add(1)
		})
		xt.Equal(t, int32(1), cnt.Load())
	})

	t.Run("case 7 ctxerr nil", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.GoCtxErr(context.Background(), func(ctx context.Context) error {
			cnt.Add(1)
			return nil
		})
		xt.NoError(t, wg.Wait())
		xt.Equal(t, int32(1), cnt.Load())

		wg.GoCtxErr(context.Background(), func(ctx context.Context) error {
			cnt.Add(1)
			return nil
		})
		xt.Equal(t, int32(1), cnt.Load())
	})

	t.Run("case 8 ctxerr err", func(t *testing.T) {
		var wg xsync.WaitFirst
		var cnt atomic.Int32
		wg.GoCtxErr(context.Background(), func(ctx context.Context) error {
			cnt.Add(1)
			return io.EOF
		})
		xt.Error(t, wg.Wait())
		xt.Equal(t, int32(1), cnt.Load())

		wg.GoCtxErr(context.Background(), func(ctx context.Context) error {
			cnt.Add(1)
			return io.EOF
		})
		xt.Equal(t, int32(1), cnt.Load())
	})
}
