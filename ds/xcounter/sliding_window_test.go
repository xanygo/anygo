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
		for i := 0; i < 10; i++ {
			ct.Incr()
		}
		xt.Equal(t, 10, ct.Count(time.Second))
		xt.Equal(t, 10, ct.Count(10*time.Second))
		xt.Equal(t, 10, ct.WindowTotal())

		time.Sleep(2 * time.Second)
		xt.Equal(t, 0, ct.Count(time.Second))
		xt.Equal(t, 10, ct.WindowTotal())
		time.Sleep(time.Hour)
		xt.Equal(t, 0, ct.CountWindow())
		xt.Equal(t, 0, ct.WindowTotal())

		ct.Incr()
		xt.Equal(t, 1, ct.Count(time.Second))
		xt.Equal(t, 1, ct.WindowTotal())
	})
}
