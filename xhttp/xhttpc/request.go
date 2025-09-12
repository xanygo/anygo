//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ xrpc.Request = (*Request)(nil)

type Request struct {
	API    string // APIName
	Method string
	Path   string
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
	proxy := xoption.Proxy(opt)
	if proxy != nil {
		tr := &httpProxyTransporter{
			conn:   node,
			config: proxy,
			req:    req,
			opt:    opt,
		}
		return tr.write(ctx)
	}
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

func (r *Request) OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader {
	opt := xoption.NewSimple()
	hc := xservice.OptHTTP(rd)
	if hc.HTTPS {
		tc1 := xoption.TLSConfig(rd).Clone()
		if tc1.ServerName == "" {
			tc1.ServerName = hc.Host
		}
		xoption.SetTLSConfig(opt, tc1)
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
	proxy := xoption.Proxy(opt)
	if proxy != nil {
		tr := &httpProxyTransporter{
			conn:   node,
			config: proxy,
			req:    req,
			opt:    opt,
		}
		return tr.write(ctx)
	}
	return req.Write(node.Conn)
}

func (h *NativeRequest) balancer(opt xoption.Reader) xbalance.Reader {
	host := h.Request.URL.Hostname()
	if host == xservice.DummyAddress {
		return nil
	}
	proxy := xoption.Proxy(opt)
	if proxy != nil {
		return nil
	}
	hostPort := getHostPort(h.Request)
	return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", hostPort))
}

func (h *NativeRequest) OptionReader(ctx context.Context, opt xoption.Reader) xoption.Reader {
	mp := xoption.NewSimple()
	if ap := h.balancer(opt); ap != nil {
		xbalance.OptSetReader(mp, ap)
	}
	if tc := h.tslConfig(opt); tc != nil {
		xoption.SetTLSConfig(mp, tc)
	}
	return mp
}

func (h *NativeRequest) tslConfig(rd xoption.Reader) *tls.Config {
	if !strings.EqualFold(h.Request.URL.Scheme, "https") {
		return nil
	}
	proxy := xoption.Proxy(rd)
	if proxy != nil {
		return nil
	}
	return getTlsConfig(h.Request, rd)
}

func setHTTPRequestUA(req *http.Request) {
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", "anygo-xrpc/1.0")
	}
}

func setHTTPRequestLogID(ctx context.Context, req *http.Request) {
	logID := xlog.FindLogID(ctx)
	if logID != "" {
		req.Header.Set("X-Log-ID", logID)
	}
}

type httpProxyTransporter struct {
	conn   *xnet.ConnNode
	config *xoption.ProxyConfig
	req    *http.Request
	opt    xoption.Reader
}

func getHostPort(req *http.Request) string {
	serverName := req.URL.Hostname()
	port := req.URL.Port()
	if port != "" {
		return net.JoinHostPort(serverName, port)
	}
	if req.URL.Scheme == "https" {
		port = "443"
	} else {
		port = "80"
	}
	return net.JoinHostPort(serverName, port)
}

// getProxyRequest 创建和 proxy 交互的首个 Request
func (p *httpProxyTransporter) getProxyRequest(ctx context.Context) (*http.Request, error) {
	hostPort := getHostPort(p.req)

	cr, err := http.NewRequestWithContext(ctx, http.MethodConnect, "http://"+hostPort, nil)
	if err != nil {
		return nil, err
	}
	cr.Host = hostPort
	cr.Header.Set("Proxy-Connection", "keep-alive")
	if p.config.Username != "" {
		switch p.config.AuthType {
		case "", "Basic":
			code := base64.StdEncoding.EncodeToString([]byte(p.config.Username + ":" + p.config.Password))
			cr.Header.Set("Proxy-Authorization", "Basic "+code)
		}
	}
	return cr, nil
}

func (p *httpProxyTransporter) write(ctx context.Context) error {
	isHTTPS := p.req.URL.Scheme == "https"
	if !isHTTPS {
		return p.req.WriteProxy(p.conn.Conn)
	}

	cr, err := p.getProxyRequest(ctx)
	if err != nil {
		return err
	}

	err = cr.Write(p.conn.Conn)
	if err != nil {
		return nil
	}
	bio := bufio.NewReader(p.conn.Conn)
	resp, err := http.ReadResponse(bio, nil)
	if err != nil {
		return err
	}
	// 代理服务器应该响应：HTTP/1.1 200 Connection Established
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %d  %s", resp.StatusCode, resp.Status)
	}
	tc := getTlsConfig(p.req, p.opt)

	// 被代理的不是 HTTPS 地址，直接发送请求
	if tc == nil {
		return p.req.Write(p.conn.Conn)
	}

	// 被代理的是 HTTPS 地址，创建加密链接，通过加密链接发送 Request、读取 Response
	pc := tls.Client(p.conn.Conn, tc)
	err = pc.HandshakeContext(ctx)
	if err != nil {
		return err
	}
	p.conn.TlsConn = pc
	return p.req.Write(pc)
}

func getTlsConfig(req *http.Request, opt xoption.Reader) *tls.Config {
	if !strings.EqualFold(req.URL.Scheme, "https") {
		return nil
	}
	serverName := req.URL.Hostname()
	tc := xoption.TLSConfig(opt)
	if tc != nil {
		tc = tc.Clone()
		if serverName != "" {
			tc.ServerName = serverName
		}
	} else {
		tc = &tls.Config{
			ServerName: serverName,
		}
	}
	return tc
}
