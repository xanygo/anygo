//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package xpp

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fsgo/fst"
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
			fst.Error(t, err)
			fst.Nil(t, fn)
		}

		time.Sleep(3 * time.Millisecond)

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			fn, err := c.WaitContext(ctx)
			fst.NoError(t, err)
			fst.NotNil(t, fn)
			fn()
		}
	})
	t.Run("limit 10", func(t *testing.T) {
		c := NewConcLimiter(10)
		start := time.Now()
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn := c.Wait()
				time.AfterFunc(time.Millisecond, func() {
					fn()
					fn() // 可重复调用
				})
			}()
		}
		wg.Wait()
		cost := time.Since(start)
		fst.GreaterOrEqual(t, int(cost/time.Millisecond), 10)
	})
}
