//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-09

package xpp

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestCooldownRunner_Run(t *testing.T) {
	var runner CooldownRunner
	synctest.Test(t, func(t *testing.T) {
		var num atomic.Int32
		for range 1000 {
			runner.Run(time.Second, func() {
				num.Add(1)
			})
			time.Sleep(100 * time.Millisecond)
		}
		synctest.Wait()
		xt.Greater(t, num.Load(), 90)
		xt.Less(t, num.Load(), 100)
	})
}
