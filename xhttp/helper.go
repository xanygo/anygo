//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xhttp

import (
	"net"
	"net/http"
	"strings"
)

func IsAjax(r *http.Request) bool {
	switch r.Header.Get("X-Requested-With") {
	case "XMLHttpRequest",
		"Fetch":
		return true
	default:
		return false
	}
}

func ClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		arr := strings.Split(xff, ",")
		if len(arr) > 0 {
			return arr[0]
		}
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
