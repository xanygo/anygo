//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xservice

import "github.com/xanygo/anygo/xoption"

var xOptKeyHTTP = xoption.NewKey("HTTP")

func SetXOptHTTP(opt xoption.Writer, val HTTPOption) {
	opt.Set(xOptKeyHTTP, val)
}

func OptHTTP(opt xoption.Reader) HTTPOption {
	return xoption.GetAsDefault[HTTPOption](opt, xOptKeyHTTP, HTTPOption{})
}
