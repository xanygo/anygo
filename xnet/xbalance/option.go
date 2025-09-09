//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-08

package xbalance

import "github.com/xanygo/anygo/xoption"

var OptionReaderKey = xoption.NewKey("balancer.Reader")

func OptSetReader(opt xoption.Writer, b Reader) {
	opt.Set(OptionReaderKey, b)
}

func OptReader(opt xoption.Reader) Reader {
	return xoption.GetAsDefault[Reader](opt, OptionReaderKey, nil)
}
