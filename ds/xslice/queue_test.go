//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-14

package xslice_test

import (
	"testing"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/xt"
)

func TestQueue_Push(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var q xslice.Queue[int]
		for i := 0; i < 10; i++ {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, 10, q.Len())

		for i := 0; i < 10; i++ {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, i, got)
		}
		xt.Equal(t, 0, q.Len())

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, 0, got)

		for i := 0; i < 10; i++ {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, 2, q.Discard(2))
		xt.Equal(t, 8, q.Len())
		xt.Equal(t, 8, q.Discard(10))
		xt.Equal(t, 0, q.Len())
		xt.Equal(t, 0, q.Discard(10))
		xt.Equal(t, 0, q.Len())
	})

	t.Run("case 2", func(t *testing.T) {
		q := &xslice.Queue[int]{
			Capacity: 3,
		}
		for i := 0; i < 3; i++ {
			xt.True(t, q.Push(i))
		}
		xt.False(t, q.Push(4))

		for i := 0; i < 3; i++ {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, i, got)
		}

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, 0, got)
	})
}

func TestSyncQueue_Push(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var q xslice.SyncQueue[int]
		for i := 0; i < 10; i++ {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, 10, q.Len())

		for i := 0; i < 10; i++ {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, i, got)
		}

		xt.Equal(t, 0, q.Len())

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, 0, got)

		for i := 0; i < 10; i++ {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, 2, q.Discard(2))
		xt.Equal(t, 8, q.Len())
		xt.Equal(t, 8, q.Discard(10))
		xt.Equal(t, 0, q.Len())
		xt.Equal(t, 0, q.Discard(10))
		xt.Equal(t, 0, q.Len())
	})

	t.Run("case 2", func(t *testing.T) {
		q := &xslice.SyncQueue[int]{
			Capacity: 3,
		}
		for i := 0; i < 3; i++ {
			xt.True(t, q.Push(i))
		}
		xt.False(t, q.Push(4))

		for i := 0; i < 3; i++ {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, i, got)
		}

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, 0, got)
	})
}
