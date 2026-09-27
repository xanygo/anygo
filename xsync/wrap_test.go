package xsync

import (
	"fmt"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestEvery(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var num int
		fn1 := Every(func() int { num++; return num }, time.Minute)
		for i := 0; i < 10; i++ {
			val := fn1()
			xt.Equal(t, val, 1)
		}
		time.Sleep(time.Minute)
		for i := 0; i < 10; i++ {
			val := fn1()
			xt.Equal(t, val, 2)
		}
	})
}

func TestEveryErr(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var num int
		fn1 := EveryErr(func() (int, error) {
			num++
			return num, fmt.Errorf("num=%d", num)
		}, time.Minute)
		for i := 0; i < 10; i++ {
			val, err := fn1()
			xt.Equal(t, val, 1)
			xt.ErrorContains(t, err, "num=1")
		}
		time.Sleep(time.Minute)
		for i := 0; i < 10; i++ {
			val, err := fn1()
			xt.Equal(t, val, 2)
			xt.ErrorContains(t, err, "num=2")
		}
	})
}

func TestDebounce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var num atomic.Int64
		fn1 := Debounce(func() {
			num.Add(1)
		}, time.Minute)
		for i := 0; i < 10; i++ {
			fn1()
			xt.Equal(t, num.Load(), 0)
		}
		time.Sleep(time.Minute + time.Second)
		for i := 0; i < 10; i++ {
			fn1()
			xt.Equal(t, num.Load(), 1)
		}
	})
}
