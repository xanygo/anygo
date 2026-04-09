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
		for i := range 10 {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, q.Len(), 10)

		for i := range 10 {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, got, i)
		}
		xt.Equal(t, q.Len(), 0)

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, got, 0)

		for i := range 10 {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, q.Discard(2), 2)
		xt.Equal(t, q.Len(), 8)
		xt.Equal(t, q.Discard(10), 8)
		xt.Equal(t, q.Len(), 0)
		xt.Equal(t, q.Discard(10), 0)
		xt.Equal(t, q.Len(), 0)
	})

	t.Run("case 2", func(t *testing.T) {
		q := &xslice.Queue[int]{
			Capacity: 3,
		}
		for i := range 3 {
			xt.True(t, q.Push(i))
		}
		xt.False(t, q.Push(4))

		for i := range 3 {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, got, i)
		}

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, got, 0)
	})
}

func TestSyncQueue_Push(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var q xslice.SyncQueue[int]
		for i := range 10 {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, q.Len(), 10)

		for i := range 10 {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, got, i)
		}

		xt.Equal(t, q.Len(), 0)

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, got, 0)

		for i := range 10 {
			xt.True(t, q.Push(i))
		}
		xt.Equal(t, q.Discard(2), 2)
		xt.Equal(t, q.Len(), 8)
		xt.Equal(t, q.Discard(10), 8)
		xt.Equal(t, q.Len(), 0)
		xt.Equal(t, q.Discard(10), 0)
		xt.Equal(t, q.Len(), 0)
	})

	t.Run("case 2", func(t *testing.T) {
		q := &xslice.SyncQueue[int]{
			Capacity: 3,
		}
		for i := range 3 {
			xt.True(t, q.Push(i))
		}
		xt.False(t, q.Push(4))

		for i := range 3 {
			got, ok := q.Pop()
			xt.True(t, ok)
			xt.Equal(t, got, i)
		}

		got, ok := q.Pop()
		xt.False(t, ok)
		xt.Equal(t, got, 0)
	})
}
