//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-23

package xhandler

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/xanygo/anygo/xlog"
)

// AntiTheft 使用 refer 信息判断请求来源是否有效
type AntiTheft struct {
	// AllowDomain  有效的域名，如 example.com,则允许此域名以及所其子域名来源的 refer
	AllowDomain []string

	xlog.WithLogger

	// Forbidden 当判断无效请求时，的回调 handler，可选
	Forbidden http.Handler

	total atomic.Int64
	fail  atomic.Int64
}

func (a *AntiTheft) Next(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.total.Add(1)
		if !a.check(r) {
			a.total.Add(1)
			if a.Forbidden == nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			} else {
				a.Forbidden.ServeHTTP(w, r)
			}
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func (a *AntiTheft) Status() map[string]any {
	return map[string]any{
		"Total": a.total.Load(),
		"Fail":  a.fail.Load(),
	}
}

func (a *AntiTheft) check(r *http.Request) bool {
	refer := r.Referer()
	if refer == "" {
		return true
	}
	u, err := url.Parse(refer)
	if err != nil {
		a.AutoLogger().Warn(r.Context(), "AntiTheft, parser refer failed",
			xlog.String("refer", refer),
			xlog.ErrorAttr("error", err),
		)
		return false
	}
	referHost := u.Hostname()
	// 请求的 HOST 和 refer 是同一个 host:port
	if r.Host == u.Host || r.Host == referHost || strings.HasSuffix(referHost, "."+r.Host) {
		return true
	}
	result := a.hostAllow(referHost)
	if !result {
		a.AutoLogger().Warn(r.Context(), "AntiTheft Forbidden",
			xlog.String("refer", refer),
			xlog.String("referHosts", referHost),
			xlog.String("host", r.Host),
		)
	}
	return result
}

func (a *AntiTheft) hostAllow(host string) bool {
	ip := net.ParseIP(host)
	if ip != nil && (ip.IsPrivate() || ip.IsLoopback()) {
		return true
	}
	for _, domain := range a.AllowDomain {
		if domain == host || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}
