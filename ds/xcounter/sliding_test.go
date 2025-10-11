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
