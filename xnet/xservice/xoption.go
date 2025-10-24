//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xservice

import (
	"github.com/xanygo/anygo/ds/xoption"
)

var (
	xOptKeyHTTP  = xoption.NewKey("HTTP")
	xOptConnPool = xoption.NewKey("ConnPool")
)

func SetOptHTTP(opt xoption.Writer, val HTTPPart) {
	opt.Set(xOptKeyHTTP, val)
}

func OptHTTP(opt xoption.Reader) HTTPPart {
	return xoption.GetAsDefault[HTTPPart](opt, xOptKeyHTTP, HTTPPart{})
}

func SetOptConnPool(opt xoption.Writer, val *ConnPoolPart) {
	opt.Set(xOptConnPool, val)
}

func OptConnPool(opt xoption.Reader) *ConnPoolPart {
	return xoption.GetAsDefault[*ConnPoolPart](opt, xOptConnPool, nil)
}
