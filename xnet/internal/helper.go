//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-28

package internal

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ParseIPZone parses s as an IP address, return it and its associated zone
// identifier (IPv6 only).
func ParseIPZone(s string) (net.IP, string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return net.ParseIP(s), ""
		case ':':
			return parseIPv6Zone(s)
		}
	}
	return nil, ""
}

// parseIPv6Zone parses s as a literal IPv6 address and its associated zone
// identifier which is described in RFC 4007.
func parseIPv6Zone(s string) (net.IP, string) {
	s, zone := splitHostZone(s)
	return net.ParseIP(s), zone
}

func splitHostZone(s string) (host, zone string) {
	// The IPv6 scoped addressing zone identifier starts after the
	// last percent sign.
	if i := last(s, '%'); i > 0 {
		host, zone = s[:i], s[i+1:]
	} else {
		host = s
	}
	return
}

func last(s string, b byte) int {
	i := len(s)
	for i--; i >= 0; i-- {
		if s[i] == b {
			break
		}
	}
	return i
}

func HostPortFromURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port, err = SchemePort(u.Scheme)
		if err != nil {
			return "", err
		}
	}
	return net.JoinHostPort(host, port), nil
}

// 目前框架中会用到的
var schemePort = map[string]string{
	"http":   "80",
	"https":  "443",
	"ws":     "80",
	"wss":    "443",
	"socks5": "1080",
	"socks4": "1080",
	// "ftp":    "21",
	// "sftp":   "22",
}

// SchemePort 获取这种 Schema 默认的端口号
func SchemePort(scheme string) (string, error) {
	key := strings.ToLower(scheme)
	port, ok := schemePort[key]
	if ok {
		return port, nil
	}
	return "", fmt.Errorf("no default port for unkonwn schema %q", scheme)
}
