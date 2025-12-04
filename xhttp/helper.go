//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xhttp

import (
	"net"
	"net/http"
	"strings"

	"github.com/xanygo/anygo/xnet/trustip"
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

// ClientIP 获取用户真实 IP
//
// 会先拿 Request.RemoteAddr，使用 trustip.IsTrusted 来判断上游来源是可信的，
//
//	若不可信，直接返回 RemoteAddr。
//	若可信，则尝试从 header： X-Real-IP、X-Forwarded-For 中读取，若没有值则最终使用 RemoteAddr。
//
// 所以，若是前端部署了 Nginx 等 LB，应确保 LB 能被 trustip 判断为是可行的，并让 LB 透传 X-Real-IP 字段作为用户真实 IP
func ClientIP(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	if !trustip.IsTrusted(ip) {
		return host
	}

	if cp := r.Header.Get("X-Real-IP"); cp != "" {
		pp := net.ParseIP(cp)
		if pp != nil {
			return cp
		}
	}
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		arr := strings.Split(xff, ",")
		if len(arr) > 0 {
			pp := net.ParseIP(arr[0])
			if pp != nil {
				return arr[0]
			}
		}
	}
	return host
}
