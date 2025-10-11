//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-07

package xcounter

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/fsgo/fst"
)

func TestSlidingDual_IncrN(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		wd := NewSlidingDual(time.Hour, time.Second)
		wd.IncrN(1, 2)
		fst.Equal(t, 3, wd.Total())
		fst.Equal(t, 1, wd.TotalSuccess())
		fst.Equal(t, 2, wd.TotalFailure())
		time.Sleep(time.Hour)
		fst.Equal(t, 0, wd.Total())
		fst.Equal(t, 0, wd.TotalSuccess())
		fst.Equal(t, 0, wd.TotalFailure())
	})
}
