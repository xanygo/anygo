//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-15

package xhttp

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xhttp/trustheader"
)

// IsAjax 判断请求是否是 AJAX 请求
//
// 判断逻辑：
//
//  1. Header["Hx-Request"] == "true", htmx 会发送该字段
//
//  2. Header["X-Requested-With"] 的值为 "XMLHttpRequest" 或 "Fetch"
//
// JavaScript Fetch API 需要通过如下方式添加 Header 标志：
//
//	fetch('/api/data', {
//		headers: {
//			'X-Requested-With': 'Fetch'
//		}
//	});
func IsAjax(r *http.Request) bool {
	hx := r.Header.Get("Hx-Request")
	if hx == "true" {
		return true
	}
	switch r.Header.Get("X-Requested-With") {
	case "XMLHttpRequest", "Fetch":
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

// RequestScheme 读取用户发起请求的  scheme，一般会返回 https 或者  http。
//
// 解析规则:
//
//  1. 若  X-Forwarded-Proto 是可信 header，并且有值，则返回。
//  2. 若 Cf-Visitor 是可信 header（cloudflare 会传递 Header Cf-Visitor: '{"scheme":"https"}'），并且可解析出 scheme 值，则返回。
//  3. 若 *http.Request.TLS != nil ，则返回 https，否则返回 http
func RequestScheme(r *http.Request) string {
	if v, ok := trustheader.Get(r.Header, "X-Forwarded-Proto"); ok {
		return v
	}

	// cloudflare: Cf-Visitor: '{"scheme":"https"}'
	if str, ok := trustheader.Get(r.Header, "Cf-Visitor"); ok && strings.Contains(str, "scheme") {
		var cf cfVisitor
		if err := xcodec.Unmarshal(xcodec.JSON, []byte(str), &cf); err == nil && cf.Scheme != "" {
			return cf.Scheme
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// RequestPort 读取用户发起请求的端口号
//
// 解析规则：
//
//  1. 若 X-Forwarded-Port 是可行 header,并且值有效，则返回
//  2. 若 Cf-Visitor 是可信 header，并且可解析出 scheme 值，则依据 scheme 返回 443 或者 80
//  3. 从 Host 解析，若有端口值则使用
//  4. 若请求的 TLS 不为 nil，则返回 443，否则返回 80
func RequestPort(r *http.Request) int {
	if v, ok := trustheader.Get(r.Header, "X-Forwarded-Port"); ok {
		if port, _ := strconv.ParseUint(v, 10, 16); port > 0 {
			return int(port)
		}
	}

	// cloudflare: Cf-Visitor: '{"scheme":"https"}'
	if str, ok := trustheader.Get(r.Header, "Cf-Visitor"); ok {
		var cf cfVisitor
		if err := xcodec.Unmarshal(xcodec.JSON, []byte(str), &cf); err == nil {
			switch cf.Scheme {
			case "https":
				return 443
			case "http":
				return 80
			}
		}
	}

	if _, port, err := net.SplitHostPort(r.Host); err == nil {
		if n, err := strconv.ParseUint(port, 10, 16); err == nil && n > 0 {
			return int(n)
		}
	}

	if r.TLS != nil {
		return 443
	}
	return 80
}

// RequestOrigin 返回用户发起请求的基础地址，格式为 schema://host:port
func RequestOrigin(r *http.Request) string {
	scheme := RequestScheme(r)
	port := RequestPort(r)

	host := r.Host
	// 去除 Host 中可能存在的端口
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return scheme + "://" + host
	}

	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(port))
}
