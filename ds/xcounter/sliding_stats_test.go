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
		xt.Equal(t, 3, wd.WindowTotal())
		xt.Equal(t, 1, wd.WindowSuccess())
		xt.Equal(t, 2, wd.WindowFailure())

		time.Sleep(time.Hour)

		xt.Equal(t, 0, wd.WindowTotal())
		xt.Equal(t, 0, wd.WindowSuccess())
		xt.Equal(t, 0, wd.WindowFailure())

		xt.Equal(t, 3, wd.LifetimeTotal())
		xt.Equal(t, 1, wd.LifetimeSuccess())
		xt.Equal(t, 2, wd.LifetimeFailure())
	})
}
