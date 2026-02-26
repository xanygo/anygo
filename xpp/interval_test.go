// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/31

package xpp_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/xpp"
	"github.com/xanygo/anygo/xt"
)

func TestInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		it := &xpp.Interval{}
		defer it.Stop()
		var num int32
		it.Add(func() {
			atomic.AddInt32(&num, 1)
		})
		var f1 int32
		it.Add(func() {
			if it.Running() {
				atomic.AddInt32(&f1, 1)
			}
		})
		var f2 int32
		var wg2 sync.WaitGroup
		for range 2 {
			wg2.Add(1)
			it.Add(func() {
				defer wg2.Done()
				select {
				case <-it.Done():
					atomic.AddInt32(&f2, 1)
				case <-time.After(20 * time.Millisecond):
					return
				}
			})
		}
		it.Start(30 * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		it.Stop()
		wg2.Wait()

		it.Reset(time.Millisecond)
		xt.Equal(t, int32(1), atomic.LoadInt32(&num))
		xt.Equal(t, int32(1), atomic.LoadInt32(&f1))
		xt.Equal(t, int32(2), atomic.LoadInt32(&f2))
	})
}

func TestInterval2(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		it := &xpp.Interval{}
		var num atomic.Int64
		it.Add(func() {
			num.Add(1)
			panic("hello")
		})
		it.Add(func() {
			num.Add(3)
			<-it.Done()
			num.Add(5)
		})
		it.Start(time.Millisecond)
		time.Sleep(time.Millisecond / 2)
		it.Stop()
		time.Sleep(time.Millisecond / 2)
		xt.Equal(t, int64(9), num.Load())
	})
}
