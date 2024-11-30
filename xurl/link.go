//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-27

package xurl

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

func NewLink(base *url.URL, param string) string {
	query := base.Query()
	p, _ := url.ParseQuery(param)
	for name, values := range p {
		value := values[0]
		if value == "" {
			query.Del(name)
		} else {
			query.Set(name, value)
		}
	}
	if len(query) == 0 {
		return base.Path
	}
	return base.Path + "?" + query.Encode()
}

var portMap = map[string]uint16{
	"ftp":    21,
	"ssh":    22,
	"sftp":   22,
	"dns":    53,
	"http":   80,
	"pop3":   110,
	"https":  443,
	"ftps":   990,
	"socks5": 1080,
}
var errEmptyHost = errors.New("empty host")

// HostPort 解析出 url 地址中的 Host 和 Port
func HostPort(u *url.URL) (host string, port uint16, err error) {
	log.Println(u.Hostname(), u.Port())
	host = u.Host
	var portStr string

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, portStr = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	if host == "" {
		return "", 0, errEmptyHost
	}
	if portStr != "" {
		num, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return host, 0, err
		}
		return host, uint16(num), err
	}
	port = portMap[u.Scheme]
	if port > 0 {
		return host, port, nil
	}
	return host, 0, fmt.Errorf("cannot get port by scheme: %s", u.Scheme)
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
