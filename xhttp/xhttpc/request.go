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
	Method string
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
	return "Request:" + r.APIName()
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
	setHTTPRequestUA(req)
	setHTTPRequestLogID(ctx, req)
	return req.Write(node.Conn)
}

func (r *Request) getURL(so xoption.Reader, address string) (string, error) {
	opt := xservice.OptHTTP(so)
	var scheme string = "http"
	if opt.HTTPS {
		scheme = "https"
	}
	u, err := url.Parse(r.Path)
	if err != nil {
		return "", err
	}
	if u.Host == xservice.DummyAddress {
		u.Host = ""
	}

	u.Scheme = scheme
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
	return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort))
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
	if proxy := xproxy.OptConfig(opt); proxy != nil {
		hostPort := getHostPort(u)
		xoption.SetTargetAddress(opt, hostPort)
	}

	hc := xservice.OptHTTP(rd)
	if hc.HTTPS || r.HTTPS {
		tc := mustGetTslConfig(u, opt)
		xoption.SetTLSConfig(opt, tc)
	}
	if opt.Empty() {
		return nil
	}
	return opt
}

var _ xrpc.Request = (*NativeRequest)(nil)

type NativeRequest struct {
	API     string
	Request *http.Request
}

func (h *NativeRequest) Protocol() string {
	return "HTTP"
}

func (h *NativeRequest) String() string {
	return "HTTPRequestNative:" + h.APIName()
}

func (h *NativeRequest) APIName() string {
	if h.API != "" {
		return h.API
	}
	return h.Request.URL.Path
}

func (h *NativeRequest) WriteTo(ctx context.Context, node *xnet.ConnNode, opt xoption.Reader) error {
	req := h.Request.WithContext(ctx)
	if req.Host == xservice.DummyAddress {
		req.Host = ""
	}
	setHTTPRequestUA(req)
	setHTTPRequestLogID(ctx, req)
	return req.Write(node.Conn)
}

func (h *NativeRequest) balancer(opt xoption.Reader) xbalance.Reader {
	host := h.Request.URL.Hostname()
	if host == xservice.Dummy {
		return nil
	}
	proxy := xproxy.OptConfig(opt)
	if proxy != nil {
		// 有代理的情况下，应该连接代理
		return nil
	}
	hostPort := getHostPort(h.Request.URL)
	return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort))
}

func (h *NativeRequest) OptionReader(ctx context.Context, opt xoption.Reader) xoption.Reader {
	mp := xoption.NewSimple()
	if ap := h.balancer(opt); ap != nil {
		xbalance.OptSetReader(mp, ap)
	}

	if proxy := xproxy.OptConfig(opt); proxy != nil {
		hostPort := getHostPort(h.Request.URL)
		xoption.SetTargetAddress(mp, hostPort)
	}
	if tc := h.tslConfig(opt); tc != nil {
		xoption.SetTLSConfig(mp, tc)
	}
	return mp
}

func (h *NativeRequest) tslConfig(opt xoption.Reader) *tls.Config {
	if !strings.EqualFold(h.Request.URL.Scheme, "https") {
		return nil
	}
	return mustGetTslConfig(h.Request.URL, opt)
}

func mustGetTslConfig(u *url.URL, opt xoption.Reader) *tls.Config {
	serverName := u.Host
	if serverName == xservice.Dummy {
		serverName = ""
	}
	tc := xoption.GetTLSConfig(opt)
	if tc != nil {
		tc = tc.Clone()
	} else {
		tc = &tls.Config{}
	}
	if serverName != "" {
		tc.ServerName = serverName
	} else {
		tc.InsecureSkipVerify = true
	}
	return tc
}

func setHTTPRequestUA(req *http.Request) {
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", xnet.UserAgent)
	}
}

func setHTTPRequestLogID(ctx context.Context, req *http.Request) {
	logID := xlog.FindLogID(ctx)
	if logID != "" {
		req.Header.Set("X-Log-ID", logID)
	}
}

func getHostPort(u *url.URL) string {
	// 此处不需要考虑 Hostname 为 dummy 的情况
	serverName := u.Hostname()
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
