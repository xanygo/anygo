// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package xsync

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestTimeStamp(t *testing.T) {
	var a TimeStamp
	got1 := a.Load()
	xt.True(t, got1.IsZero())
	now := time.Now()
	a.Store(now)
	got2 := a.Load()
	xt.Equal(t, now.UnixNano(), got2.UnixNano())

	t2 := now.Add(time.Second)
	xt.True(t, a.Before(t2))

	t3 := now.Add(-1 * time.Second)
	xt.True(t, a.After(t3))

	xt.Equal(t, time.Second, a.Sub(t3))

	xt.GreaterOrEqual(t, a.Since(time.Now()), time.Duration(0))
}

func BenchmarkTimeStamp_Load(b *testing.B) {
	var a TimeStamp
	a.Store(time.Now())
	b.ResetTimer()
	var tm time.Time
	for i := 0; i < b.N; i++ {
		tm = a.Load()
	}
	_ = tm
}

func TestInterval_Allow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var it Interval
		xt.True(t, it.Allow(time.Minute))
		xt.False(t, it.Allow(time.Minute))
		time.Sleep(time.Minute + 1)
		xt.True(t, it.Allow(time.Minute))
	})
}
