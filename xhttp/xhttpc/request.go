//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xproxy"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ xrpc.Request = (*Request)(nil)

type Request struct {
	API    string // APIName
	Method string // 请求方法，可选，默认为 http.MethodGet
	Path   string
	HTTPS  bool
	Query  url.Values
	Header http.Header
	Body   io.Reader
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

func (r *Request) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	api, err := r.getURL(opt, node.Addr.HostPort)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, r.getMethod(), api, r.Body)
	if err != nil {
		return err
	}
	if req.Host == xservice.Dummy {
		req.Host = ""
	}
	setHeader(ctx, req, opt)
	return req.Write(node.Outer())
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
	if u.Host == "" || u.Hostname() == xservice.Dummy {
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
}

func (h *NativeRequest) Protocol() string {
	return "HTTP"
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

func (h *NativeRequest) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	req := h.Request.WithContext(ctx)
	if req.Host == xservice.Dummy {
		req.Host = ""
	}
	if req.Host == "" {
		req.Host = node.Addr.Host()
	}
	setHeader(ctx, req, opt)
	return req.Write(node.Outer())
}

func (h *NativeRequest) balancer(opt xoption.Reader) xbalance.Reader {
	host := h.Request.URL.Hostname()
	if host == xservice.Dummy {
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
		req.Header[k] = v
	}
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", xnet.UserAgent)
	}
	logID := xlog.FindLogID(ctx)
	if logID != "" {
		req.Header.Set("X-Log-ID", logID)
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
