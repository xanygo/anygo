//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

package xoption

import "github.com/xanygo/anygo/xnet/xpolicy"

func SetRetry(opt Writer, retry int) {
	retry = max(0, retry)
	opt.Set(KeyRetry, retry)
}

func Retry(opt Reader) int {
	return Int(opt, KeyRetry, DefaultRetry)
}

func SetRetryPolicy(opt Writer, policy *xpolicy.Retry) {
	opt.Set(KeyRetryPolicy, policy)
}

func RetryPolicy(opt Reader) *xpolicy.Retry {
	val, ok := GetAs[*xpolicy.Retry](opt, KeyRetryPolicy)
	if ok && val != nil {
		return val
	}
	return xpolicy.DefaultRetry()
}
