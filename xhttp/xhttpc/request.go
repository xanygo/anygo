//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xpolicy"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
)

var defaultUa = &xsync.OnceInit[string]{
	New: func() string {
		return xnet.UserAgent
	},
}

func SetDefaultUserAgent(userAgent string) {
	defaultUa.Store(userAgent)
}

func DefaultUserAgent() string {
	return defaultUa.Load()
}

var _ xrpc.Request = (*Request)(nil)

type Request struct {
	API     string // APIName
	Method  string // 请求方法，可选，默认为 http.MethodGet
	Path    string
	HTTPS   bool
	Query   url.Values
	Header  http.Header
	GetBody func() (io.ReadCloser, error)

	// Idempotency 多次发送该请求，Server 端的结果是否幂等，可选
	// 当不设置的时候，会依据 Method 判断
	Idempotency xtype.TriState
}

var _ xpolicy.Idempotent = (*Request)(nil)

func (r *Request) Idempotent() bool {
	if r == nil {
		return false
	}
	if r.Idempotency.NotNull() {
		return r.Idempotency.IsTrue()
	}
	return retryableMethod(r.Method)
}

func retryableMethod(method string) bool {
	method = strings.ToUpper(method)
	switch method {
	case "", // 空等于 GET
		http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

func (r *Request) Protocol() string {
	return "HTTP"
}

func (r *Request) String() string {
	var b strings.Builder
	b.WriteString("Request: ")
	b.WriteString(r.getMethod())
	if r.HTTPS {
		b.WriteString(" https ")
	} else {
		b.WriteString("http ")
	}
	b.WriteString(r.Path)
	if len(r.Query) > 0 {
		b.WriteString(" ? ")
		b.WriteString(r.Query.Encode())
	}
	return b.String()
}

func (r *Request) APIName() string {
	if r.API != "" {
		return r.API
	}
	return r.Path
}

func (r *Request) GetMethod() string {
	return r.getMethod()
}

func (r *Request) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	api, err := r.getURL(opt, node.Addr.HostPort)
	if err != nil {
		return err
	}

	timeout := xoption.WriteTimeout(opt)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err = node.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer node.SetDeadline(time.Time{})

	req, err := http.NewRequestWithContext(ctx, r.getMethod(), api, nil)
	if err != nil {
		return err
	}
	if r.GetBody != nil {
		req.GetBody = r.GetBody
		bd, err := r.GetBody()
		if err != nil {
			return err
		}
		req.Body = bd
	}
	if req.Host == xnet.Dummy {
		req.Host = ""
	}
	setHeader(ctx, req, opt)
	err = req.Write(node.Outer())
	if err == nil {
		return nil
	}
	return fmt.Errorf("request.Write: %w", err)
}

func (r *Request) getURL(so xoption.Reader, address string) (string, error) {
	opt := xservice.OptHTTP(so)
	u, err := url.Parse(r.Path)
	if err != nil {
		return "", err
	}
	if u.Host == xnet.DummyAddress {
		u.Host = ""
	}

	u.Scheme = "http" // 不需要区分是 HTTP 还是 HTTPS
	if opt.Host != "" {
		u.Host = opt.Host
	}
	if u.Host == "" {
		u.Host = address
	}
	return u.String(), nil
}

func (r *Request) getMethod() string {
	if r.Method == "" {
		return http.MethodGet
	}
	return r.Method
}

func (r *Request) balancer(opt xoption.Reader, u *url.URL) xbalance.Reader {
	if u.Host == "" || u.Hostname() == xnet.Dummy {
		return nil
	}
	proxy := xproxy.OptConfig(opt)
	if proxy != nil {
		return nil
	}
	hostPort := getHostPort(u)
	if hostPort != "" {
		node := xnet.AddrNode{
			Addr:     xnet.NewAddr("tcp", hostPort),
			HostPort: hostPort,
		}
		return xbalance.NewStatic(node)
	}
	return nil
}

func (r *Request) tlsConfig(u *url.URL) *tls.Config {
	if !r.HTTPS {
		return nil
	}
	return &tls.Config{
		ServerName: u.Hostname(),
	}
}

func (r *Request) OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader {
	u, err := url.Parse(r.Path)
	if err != nil {
		return nil
	}
	opt := xoption.NewSimple()
	if b := r.balancer(rd, u); b != nil {
		xbalance.OptSetReader(opt, b)
	}
	if tc := r.tlsConfig(u); tc != nil {
		xoption.SetTLSConfig(opt, tc)
	}
	return opt.Value()
}

var _ xrpc.Request = (*NativeRequest)(nil)

type NativeRequest struct {
	API     string
	Request *http.Request // 必填

	// Idempotency 多次发送该请求，Server 端的结果是否幂等，可选
	// 当不设置的时候，会依据 Method 判断
	Idempotency xtype.TriState
}

func (h *NativeRequest) Protocol() string {
	return "HTTP"
}

var _ xpolicy.Idempotent = (*NativeRequest)(nil)

func (h *NativeRequest) Idempotent() bool {
	if h == nil || h.Request == nil {
		return false
	}
	if h.Idempotency.NotNull() {
		return h.Idempotency.IsTrue()
	}
	return retryableMethod(h.Request.Method)
}

func (h *NativeRequest) String() string {
	var b strings.Builder
	b.WriteString("NativeRequest: ")
	b.WriteString(h.Request.Method)
	b.WriteString(" ")
	b.WriteString(h.Request.URL.String())
	return b.String()
}

func (h *NativeRequest) APIName() string {
	if h.API != "" {
		return h.API
	}
	return h.Request.URL.Path
}

func (h *NativeRequest) GetMethod() string {
	return h.Request.Method
}

func (h *NativeRequest) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	timeout := xoption.WriteTimeout(opt)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := node.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer node.SetDeadline(time.Time{})

	req := h.Request.Clone(ctx)
	if req.Host == xnet.Dummy {
		req.Host = ""
	}
	if req.Host == "" {
		req.Host = node.Addr.Host()
	}
	if req.GetBody != nil {
		bd, err := req.GetBody()
		if err != nil {
			return err
		}
		req.Body = bd
	}
	setHeader(ctx, req, opt)
	err := req.Write(node.Outer())
	if err == nil {
		return nil
	}
	return fmt.Errorf("nativeRequest.Write: %w", err)
}

func (h *NativeRequest) balancer(opt xoption.Reader) xbalance.Reader {
	host := h.Request.URL.Hostname()
	if host == xnet.Dummy {
		return nil
	}
	if hostPort := getHostPort(h.Request.URL); hostPort != "" {
		return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort))
	}
	return nil
}

func (h *NativeRequest) tlsConfig() *tls.Config {
	if h.Request.URL.Scheme != "https" {
		return nil
	}
	return &tls.Config{
		ServerName: h.Request.URL.Hostname(),
	}
}

func (h *NativeRequest) OptionReader(ctx context.Context, opt xoption.Reader) xoption.Reader {
	mp := xoption.NewSimple()
	if ap := h.balancer(opt); ap != nil {
		xbalance.OptSetReader(mp, ap)
	}

	if tc := h.tlsConfig(); tc != nil {
		xoption.SetTLSConfig(mp, tc)
	}

	return mp.Value()
}

func setHeader(ctx context.Context, req *http.Request, opt xoption.Reader) {
	hc := xservice.OptHTTP(opt)
	if hc.Host != "" {
		req.Host = hc.Host
	}
	for k, v := range hc.Header {
		req.Header[k] = slices.Clone(v)
	}
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", DefaultUserAgent())
	}
	logID := xlog.FindLogID(ctx)
	if logID != "" {
		req.Header.Set("X-Log-ID", logID)
	}
	if attempt := xrpc.RetryCountFromContext(ctx); attempt > 0 {
		req.Header.Set("X-Retry-Count", strconv.Itoa(attempt))
	}
}

func getHostPort(u *url.URL) string {
	// 此处不需要考虑 Hostname 为 dummy 的情况
	serverName := u.Hostname()
	if serverName == "" {
		return ""
	}
	port := u.Port()
	if port != "" {
		return net.JoinHostPort(serverName, port)
	}
	if u.Scheme == "https" {
		port = "443"
	} else {
		port = "80"
	}
	return net.JoinHostPort(serverName, port)
}
