//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xhttp

import (
	"net"
	"net/http"
	"strings"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xhttp/trustheader"
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

// ClientIP 获取用户真实 IP(不包括端口号，可能是 IPV4 和 IPV6 地址)，有可能返回空
//
// 读取规则：
//  1. 若 Header 中的  Cf-Connecting-Ip 是可信并且有值，直接返回
//  2. 若 Header 中的  X-Real-IP 是可信并且有值，直接返回
//  3. 若 Header 中的  X-Forwarded-For 是可信并且有值，返回第一个不为空的
//  4. 读取 *http.Request.RemoteAddr,返回 ip 值
func ClientIP(r *http.Request) string {
	if v, ok := trustheader.Get(r.Header, "Cf-Connecting-Ip"); ok {
		return v
	}

	if v, ok := trustheader.Get(r.Header, "X-Real-IP"); ok {
		return v
	}

	if xff, ok := trustheader.Get(r.Header, "X-Forwarded-For"); ok {
		arr := strings.Split(xff, ",")
		for _, v := range arr {
			ip := strings.TrimSpace(v)
			if ip != "" {
				return ip
			}
		}
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

type cfVisitor struct {
	Scheme string `json:"scheme"`
}

// ClientScheme 从请求中读取用户发起请求的  scheme，一般会返回 https 或者  http。
//
// 解析规则:
//
//  1. 若  X-Forwarded-Proto 是可信 header，并且有值，则返回。
//  2. 若 Cf-Visitor 是可信 header（cloudflare 会传递 Header Cf-Visitor: '{"scheme":"https"}'），并且可解析出 scheme 值，则返回。
//  3. 若 *http.Request.TLS != nil ，则返回 https，否则返回 http
func ClientScheme(r *http.Request) string {
	if v, ok := trustheader.Get(r.Header, "X-Forwarded-Proto"); ok {
		return v
	}

	// cloudflare: Cf-Visitor: '{"scheme":"https"}'
	if str, ok := trustheader.Get(r.Header, "Cf-Visitor"); ok && strings.Contains(str, "scheme") {
		var cf cfVisitor
		if err := xcodec.Decode(xcodec.JSON, []byte(str), &cf); err == nil && cf.Scheme != "" {
			return cf.Scheme
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
