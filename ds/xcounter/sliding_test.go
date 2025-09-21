//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-18

package xcounter

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/fsgo/fst"
)

func TestSliding_Incr(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ct := NewSliding(time.Hour, time.Second)
		for i := 0; i < 10; i++ {
			ct.Incr()
		}
		fst.Equal(t, 10, ct.Count(time.Second))
		fst.Equal(t, 10, ct.Count(10*time.Second))
		fst.Equal(t, 10, ct.Total())

		time.Sleep(2 * time.Second)
		fst.Equal(t, 0, ct.Count(time.Second))
		fst.Equal(t, 10, ct.Total())
		time.Sleep(time.Hour)
		fst.Equal(t, 0, ct.CountWindow())
		fst.Equal(t, 0, ct.Total())

		ct.Incr()
		fst.Equal(t, 1, ct.Count(time.Second))
		fst.Equal(t, 1, ct.Total())
	})
}

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
