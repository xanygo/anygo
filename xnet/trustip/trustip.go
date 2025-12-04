//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-04

package trustip

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var defaultManager = New()

func init() {
	defaultManager.MustAdd("127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16")
	defaultManager.MustAdd("::1/128", "fc00::/7", "fe80::/10")

	const envKey = "ANYGO_TRUSTIP"

	if val := os.Getenv(envKey); val != "" {
		val = strings.ReplaceAll(val, ";", ",")
		if err := defaultManager.Set(strings.Split(val, ",")); err != nil {
			err = fmt.Errorf("parser env.%q=%q: %w", envKey, val, err)
			panic(err)
		}
	}
}

// Default 返回全局单例 Manager
func Default() *Manager {
	return defaultManager
}

func Set(cidrs []string) error {
	return Default().Set(cidrs)
}

func MustSet(cidrs []string) {
	Default().MustSet(cidrs)
}

func Add(cidrs ...string) error {
	return Default().Add(cidrs...)
}

func MustAdd(cidrs ...string) {
	Default().MustAdd(cidrs...)
}

func Remove(cidr string) error {
	return Default().Remove(cidr)
}

func RemoveIPNet(ipNet *net.IPNet) bool {
	return Default().RemoveIPNet(ipNet)
}

func List() []string {
	return Default().List()
}

func IsTrusted(ip net.IP) bool {
	return Default().IsTrusted(ip)
}
