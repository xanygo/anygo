//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xnet

import (
	"net"
	"strings"
)

const (
	// Dummy 虚拟的主机名，用于创建虚拟 service 等一些特殊的逻辑
	Dummy       = "dummy"
	dummyPrefix = Dummy + "_"

	dummyIP = "192.0.2.8"

	// DummyAddress 使用 Hostname,方便框架其他地方做一些特殊的判断
	DummyAddress = Dummy + ":80"
)

// 使用 rfc5737 中规定的 TEST-NET-1 地址中的一个
var dummyIPS = []net.IP{net.ParseIP(dummyIP)}

var errDialDummy = &net.AddrError{
	Err:  "Access to dummy IP is forbidden",
	Addr: dummyIP,
}

func IsDummyName(name string) bool {
	return name == Dummy || strings.HasPrefix(name, dummyPrefix)
}
