//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-07

package xcounter_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/ds/xcounter"
	"github.com/xanygo/anygo/xt"
)

func TestSlidingDual_IncrN(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		wd := xcounter.NewSlidingWindowStats(time.Hour, time.Second)
		wd.IncrN(1, 2)
		xt.Equal(t, wd.WindowTotal(), 3)
		xt.Equal(t, wd.WindowSuccess(), 1)
		xt.Equal(t, wd.WindowFailure(), 2)

		time.Sleep(time.Hour)

		xt.Equal(t, wd.WindowTotal(), 0)
		xt.Equal(t, wd.WindowSuccess(), 0)
		xt.Equal(t, wd.WindowFailure(), 0)

		xt.Equal(t, wd.LifetimeTotal(), 3)
		xt.Equal(t, wd.LifetimeSuccess(), 1)
		xt.Equal(t, wd.LifetimeFailure(), 2)
	})
}
