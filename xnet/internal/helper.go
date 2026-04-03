//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-28

package internal

import (
	"fmt"
	"net"
	"net/url"
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
		switch u.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return "", fmt.Errorf("invalid url %q, missing port", urlStr)
		}
	}
	return net.JoinHostPort(host, port), nil
}
