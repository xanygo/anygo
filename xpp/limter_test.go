//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package xpp

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestConcurrency(t *testing.T) {
	t.Run("limit 1", func(t *testing.T) {
		c := NewConcLimiter(1)

		done := make(chan bool)
		go func() {
			re := c.Wait()
			time.AfterFunc(3*time.Millisecond, re)
			done <- true
		}()

		<-done
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			fn, err := c.WaitContext(ctx)
			xt.Error(t, err)
			xt.Nil(t, fn)
		}

		time.Sleep(3 * time.Millisecond)

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			fn, err := c.WaitContext(ctx)
			xt.NoError(t, err)
			xt.NotNil(t, fn)
			fn()
		}
	})
	t.Run("limit 10", func(t *testing.T) {
		c := NewConcLimiter(10)
		start := time.Now()
		var wg sync.WaitGroup
		for range 100 {
			wg.Go(func() {
				fn := c.Wait()
				time.AfterFunc(time.Millisecond, func() {
					fn()
					fn() // 可重复调用
				})
			})
		}
		wg.Wait()
		cost := time.Since(start)
		xt.GreaterOrEqual(t, int(cost/time.Millisecond), 10)
	})
}
