//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package xcounter

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestSliding_Incr(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ct := NewSlidingWindow(time.Hour, time.Second)
		for range 10 {
			ct.Incr()
		}
		xt.Equal(t, ct.Count(time.Second), 10)
		xt.Equal(t, ct.Count(10*time.Second), 10)
		xt.Equal(t, ct.WindowTotal(), 10)

		time.Sleep(2 * time.Second)
		xt.Equal(t, ct.Count(time.Second), 0)
		xt.Equal(t, ct.WindowTotal(), 10)
		time.Sleep(time.Hour)
		xt.Equal(t, ct.CountWindow(), 0)
		xt.Equal(t, ct.WindowTotal(), 0)

		ct.Incr()
		xt.Equal(t, ct.Count(time.Second), 1)
		xt.Equal(t, ct.WindowTotal(), 1)
	})
}
