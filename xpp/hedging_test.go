//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-22

package xpp_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xpp"
)

func TestHedging_Run(t *testing.T) {
	t.Run("no fn 1", func(t *testing.T) {
		h1 := &xpp.Hedging[int]{
			Main: func(ctx context.Context) (int, error) {
				return 1, nil
			},
		}
		got, err := h1.Run(context.Background())
		fst.NoError(t, err)
		fst.Equal(t, 1, got)
	})
	t.Run("no fn 2", func(t *testing.T) {
		h1 := &xpp.Hedging[int]{
			Main: func(ctx context.Context) (int, error) {
				return 0, io.EOF
			},
		}
		got, err := h1.Run(context.Background())
		fst.Error(t, err)
		fst.Equal(t, 0, got)
	})
	t.Run("fn 1", func(t *testing.T) {
		h1 := &xpp.Hedging[int]{
			Main: func(ctx context.Context) (int, error) {
				select {
				case <-ctx.Done():
				case <-time.After(time.Second):
				}
				return 1, nil
			},
		}
		h1.Add(10*time.Microsecond, func(ctx context.Context) (int, error) {
			return 2, nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		got, err := h1.Run(ctx)
		fst.NoError(t, err)
		fst.Equal(t, 2, got)
	})

	t.Run("fn 2", func(t *testing.T) {
		h1 := &xpp.Hedging[int]{
			Main: func(ctx context.Context) (int, error) {
				select {
				case <-ctx.Done():
				case <-time.After(time.Second):
				}
				return 1, nil
			},
		}
		h1.Add(10*time.Microsecond, func(ctx context.Context) (int, error) {
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
			return 2, nil
		})
		h1.Add(30*time.Microsecond, func(ctx context.Context) (int, error) {
			return 3, nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		got, err := h1.Run(ctx)
		fst.NoError(t, err)
		fst.Equal(t, 3, got)
	})

	t.Run("fn 2 CallNext", func(t *testing.T) {
		h1 := &xpp.Hedging[int]{
			Main: func(ctx context.Context) (int, error) {
				select {
				case <-ctx.Done():
				case <-time.After(time.Second):
				}
				return 1, nil
			},
			CallNext: func(ctx context.Context, value int, err error) bool {
				return err != nil
			},
		}
		h1.Add(10*time.Microsecond, func(ctx context.Context) (int, error) {
			return 2, io.EOF // 触发 CallNext
		})
		h1.Add(40*time.Microsecond, func(ctx context.Context) (int, error) {
			return 3, nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		got, err := h1.Run(ctx)
		fst.NoError(t, err)
		fst.Equal(t, 3, got)
	})
	t.Run("fn panic", func(t *testing.T) {
		h1 := &xpp.Hedging[int]{
			Main: func(ctx context.Context) (int, error) {
				select {
				case <-ctx.Done():
				case <-time.After(time.Second):
				}
				return 1, nil
			},
			CallNext: func(ctx context.Context, value int, err error) bool {
				return err != nil
			},
		}
		h1.Add(10*time.Microsecond, func(ctx context.Context) (int, error) {
			panic("hello")
		})
		h1.Add(30*time.Microsecond, func(ctx context.Context) (int, error) {
			return 3, nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		got, err := h1.Run(ctx)
		fst.NoError(t, err)
		fst.Equal(t, 3, got)
	})
}
