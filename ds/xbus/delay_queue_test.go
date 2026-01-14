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
		go func() {
			for i := 1; i < 10; i++ {
				ok := q.Push(i)
				xt.True(t, ok)
			}
		}()
		for i := 1; i < 10; i++ {
			v, err := q.PopWait()
			// t.Logf("PopWait <%d: %d, %v>", i, v, err)
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
		go func() {
			for i := 1; i < 10; i++ {
				ok := q.Push(i)
				xt.True(t, ok)
			}
		}()
		now := time.Now()
		for i := 1; i < 10; i++ {
			v, err := q.PopWait()
			// t.Logf("PopWait <%d: %d, %v>", i, v, err)
			xt.NoError(t, err)
			xt.Equal(t, i, v)
			delay := time.Since(now)
			xt.GreaterOrEqual(t, delay, q.Delay)
		}
		xt.Equal(t, 0, q.Len())
	})
}
