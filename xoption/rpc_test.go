//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xoption

import (
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestSetConnectTimeout(t *testing.T) {
	doCheck := func(t *testing.T, opt Option) {
		fst.Equal(t, DefaultRetry, Retry(opt))
		SetRetry(opt, 2)
		fst.Equal(t, 2, Retry(opt))
		opt.Delete(KeyRetry)
		fst.Equal(t, DefaultRetry, Retry(opt))

		fst.Equal(t, DefaultReadTimeout, ReadTimeout(opt))
		SetReadTimeout(opt, time.Hour)
		fst.Equal(t, time.Hour, ReadTimeout(opt))

		fst.Equal(t, DefaultWriteTimeout, WriteTimeout(opt))
		SetWriteTimeout(opt, time.Minute)
		fst.Equal(t, time.Minute, WriteTimeout(opt))

		fst.Equal(t, 64*mb, MaxResponseSize(opt))
		SetMaxResponseSize(opt, 5)
		fst.Equal(t, 5, MaxResponseSize(opt))

		fst.Equal(t, "RoundRobin", Balancer(opt))
		SetBalancer(opt, "demo")
		fst.Equal(t, "demo", Balancer(opt))

		var total int
		opt.Range(func(key Key, val any) bool {
			total++
			return true
		})
		fst.Equal(t, 4, total)
	}

	opt1 := NewSimple()
	doCheck(t, opt1)

	opt2 := NewDynamic()
	doCheck(t, opt2)
}
