//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xoption

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestSetConnectTimeout(t *testing.T) {
	doCheck := func(t *testing.T, opt Option) {
		xt.Equal(t, DefaultRetry, Retry(opt))
		SetRetry(opt, 2)
		xt.Equal(t, 2, Retry(opt))
		opt.Delete(KeyRetry)
		xt.Equal(t, DefaultRetry, Retry(opt))

		xt.Equal(t, DefaultReadTimeout, ReadTimeout(opt))
		SetReadTimeout(opt, time.Hour)
		xt.Equal(t, time.Hour, ReadTimeout(opt))

		xt.Equal(t, DefaultWriteTimeout, WriteTimeout(opt))
		SetWriteTimeout(opt, time.Minute)
		xt.Equal(t, time.Minute, WriteTimeout(opt))

		xt.Equal(t, 64*mb, MaxResponseSize(opt))
		SetMaxResponseSize(opt, 5)
		xt.Equal(t, 5, MaxResponseSize(opt))

		xt.Equal(t, "RoundRobin", Balancer(opt))
		SetBalancer(opt, "demo")
		xt.Equal(t, "demo", Balancer(opt))

		var total int
		opt.Range(func(key Key, val any) bool {
			total++
			return true
		})
		xt.Equal(t, 4, total)
	}

	opt1 := NewSimple()
	doCheck(t, opt1)

	opt2 := NewDynamic()
	doCheck(t, opt2)
}
