//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-04

// Package trustip 判断请求来源是否位于可信 IP 区间
//
// 默认的 Default() 已经初始化并设置如下地址为可信范围：
//
//	IPv4: "127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"
//	IPv6: "::1/128", "fc00::/7", "fe80::/10"
//
// 同时，也可以在程序启动前，通过设置环境变量覆盖默认设置：
//
//	export ANYGO_TRUSTIP="127.0.0.1/32,10.0.0.0/8"
//
//	使用 IsTrusted(ip net.IP) 方法来判断是否在设置的区间内
package trustip
