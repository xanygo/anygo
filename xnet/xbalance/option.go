//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-04

package xbalance

import (
	xoption2 "github.com/xanygo/anygo/ds/xoption"
)

var (
	// OptKeyDownstream 用于当前直接连接的下游地址
	OptKeyDownstream = xoption2.NewKey("opt.balancer")
)

func OptSetReader(opt xoption2.Writer, b Reader) {
	opt.Set(OptKeyDownstream, b)
}

func OptReader(opt xoption2.Reader) Reader {
	return xoption2.GetAsDefault[Reader](opt, OptKeyDownstream, nil)
}
