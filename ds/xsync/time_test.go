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
	t.Run("test 1", func(t *testing.T) {
		var a TimeStamp
		got1 := a.Load()
		xt.True(t, got1.IsZero())
		xt.True(t, a.BeforeNow())

		now := time.Now()
		a.Store(now)
		got2 := a.Load()
		xt.Equal(t, got2.UnixNano(), now.UnixNano())

		t2 := now.Add(time.Second)
		xt.True(t, a.Before(t2))

		t3 := now.Add(-1 * time.Second)
		xt.True(t, a.After(t3))

		xt.Equal(t, a.Sub(t3), time.Second)

		xt.GreaterOrEqual(t, a.Since(time.Now()), time.Duration(0))
	})

	t.Run("case 2", func(t *testing.T) {
		var a TimeStamp
		xt.True(t, a.BeforeNow())

		a.Store(time.Now().Add(time.Hour))
		xt.True(t, a.AfterNow())
		xt.False(t, a.BeforeNow())
		xt.False(t, a.BeforePlus(time.Second))
		xt.True(t, a.AfterPlus(time.Second))
	})
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
