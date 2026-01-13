//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-09

package xbus_test

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xbus"
	"github.com/xanygo/anygo/xt"
)

func TestDelayQueue(t *testing.T) {
	t.Run("no delay", func(t *testing.T) {
		var q xbus.DelayQueue[int]
		defer q.Stop()

		for i := 1; i < 10; i++ {
			q.Push(i)
		}
		for i := 1; i < 10; i++ {
			v, err := q.PopWait()
			xt.NoError(t, err)
			xt.Equal(t, i, v)
		}
		xt.Equal(t, 0, q.Len())
	})

	t.Run("delay", func(t *testing.T) {
		q := &xbus.DelayQueue[int]{
			Delay: 100 * time.Millisecond,
		}
		defer q.Stop()
		for i := 1; i < 10; i++ {
			q.Push(i)
		}
		now := time.Now()
		for i := 1; i < 10; i++ {
			v, err := q.PopWait()
			xt.NoError(t, err)
			xt.Equal(t, i, v)
			delay := time.Since(now)
			xt.GreaterOrEqual(t, delay, q.Delay)
		}
		xt.Equal(t, 0, q.Len())
	})

	t.Run("delete 1", func(t *testing.T) {
		q := &xbus.DelayQueue[int]{}
		defer q.Stop()
		for i := 0; i < 10; i++ {
			q.Push(i)
		}
		xt.Equal(t, 10, q.Len())
		deleted := q.DeleteByFunc(func(v int) bool {
			return v%2 == 0
		})
		xt.Equal(t, 5, deleted)
		xt.Less(t, q.Len(), 10)
		for i := 0; i < 10; i++ {
			if v, ok := q.TryPop(); ok {
				xt.False(t, v%5 == 0)
			}
		}
	})
	t.Run("delete 2", func(t *testing.T) {
		q := &xbus.DelayQueue[int]{}
		defer q.Stop()

		for i := 0; i < 10; i++ {
			q.Push(i)
		}
		xt.Equal(t, 10, q.Len())
		deleted := q.DeleteByFunc(func(v int) bool {
			return v%5 == 0
		})
		xt.Equal(t, 2, deleted)
		xt.Less(t, q.Len(), 10)
		for i := 0; i < 10; i++ {
			if v, ok := q.TryPop(); ok {
				xt.False(t, v%5 == 0)
			}
		}
	})
}
