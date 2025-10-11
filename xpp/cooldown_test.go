//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-09

package xpp

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fsgo/fst"
)

func TestCooldownRunner_Run(t *testing.T) {
	var runner CooldownRunner
	synctest.Test(t, func(t *testing.T) {
		var num atomic.Int32
		for i := 0; i < 1000; i++ {
			runner.Run(time.Second, func() {
				num.Add(1)
			})
			time.Sleep(100 * time.Millisecond)
		}
		synctest.Wait()
		fst.Greater(t, num.Load(), 90)
		fst.Less(t, num.Load(), 100)
	})
}
