//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-04

package xbalance

import "github.com/xanygo/anygo/xoption"

var (
	// OptKeyDownstream 用于当前直接连接的下游地址
	OptKeyDownstream = xoption.NewKey("opt.balancer")
)

func OptSetReader(opt xoption.Writer, b Reader) {
	opt.Set(OptKeyDownstream, b)
}

func OptReader(opt xoption.Reader) Reader {
	return xoption.GetAsDefault[Reader](opt, OptKeyDownstream, nil)
}
