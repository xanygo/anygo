//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-08

package xbalance

import "github.com/xanygo/anygo/xoption"

var (
	// OptKeyDownstream 用于当前直接连接的下游地址
	OptKeyDownstream = xoption.NewKey("downstream.balancer")

	// OptKeyTarget 在使用代理服务的时候，用于获取实际目标服务器的 host:port
	OptKeyTarget = xoption.NewKey("target.balancer")
)

func OptSetDownstream(opt xoption.Writer, b Reader) {
	opt.Set(OptKeyDownstream, b)
}

func OptDownstream(opt xoption.Reader) Reader {
	return xoption.GetAsDefault[Reader](opt, OptKeyDownstream, nil)
}

func OptSetTarget(opt xoption.Writer, b Reader) {
	opt.Set(OptKeyTarget, b)
}

func OptTarget(opt xoption.Reader) Reader {
	return xoption.GetAsDefault[Reader](opt, OptKeyTarget, nil)
}
