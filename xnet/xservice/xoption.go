//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xservice

import (
	xoption2 "github.com/xanygo/anygo/ds/xoption"
)

var (
	xOptKeyHTTP  = xoption2.NewKey("HTTP")
	xOptConnPool = xoption2.NewKey("ConnPool")
)

func SetOptHTTP(opt xoption2.Writer, val HTTPPart) {
	opt.Set(xOptKeyHTTP, val)
}

func OptHTTP(opt xoption2.Reader) HTTPPart {
	return xoption2.GetAsDefault[HTTPPart](opt, xOptKeyHTTP, HTTPPart{})
}

func SetOptConnPool(opt xoption2.Writer, val *ConnPoolPart) {
	opt.Set(xOptConnPool, val)
}

func OptConnPool(opt xoption2.Reader) *ConnPoolPart {
	return xoption2.GetAsDefault[*ConnPoolPart](opt, xOptConnPool, nil)
}
