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
		xt.Equal(t, Retry(opt), DefaultRetry)
		SetRetry(opt, 2)
		xt.Equal(t, Retry(opt), 2)
		opt.Delete(KeyRetry)
		xt.Equal(t, Retry(opt), DefaultRetry)

		xt.Equal(t, ReadTimeout(opt), DefaultReadTimeout)
		SetReadTimeout(opt, time.Hour)
		xt.Equal(t, ReadTimeout(opt), time.Hour)

		xt.Equal(t, WriteTimeout(opt), DefaultWriteTimeout)
		SetWriteTimeout(opt, time.Minute)
		xt.Equal(t, WriteTimeout(opt), time.Minute)

		xt.Equal(t, MaxResponseSize(opt), 64*1024*1024)
		SetMaxResponseSize(opt, 5)
		xt.Equal(t, MaxResponseSize(opt), 5)

		xt.Equal(t, Balancer(opt), "RoundRobin")
		SetBalancer(opt, "demo")
		xt.Equal(t, Balancer(opt), "demo")

		var total int
		opt.Range(func(key Key, val any) bool {
			total++
			return true
		})
		xt.Equal(t, total, 4)
	}

	opt1 := NewSimple()
	doCheck(t, opt1)

	opt2 := NewDynamic()
	doCheck(t, opt2)
}
