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
	"strings"

	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xrpc"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

const HostDummy = xrpc.HostDummy

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

func (r *Request) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
	conn, ok := w.(net.Conn)
	if !ok {
		return fmt.Errorf("writer (%T) not a net.Conn", w)
	}
	api, err := r.getURL(opt, conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, r.getMethod(), api, r.Body)
	if err != nil {
		return err
	}
	setHTTPRequestUA(req)
	return req.Write(w)
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
	if u.Host == HostDummy {
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
	opt := xoption.NewMapOption()
	hc := xservice.OptHTTP(rd)
	if hc.HTTPS {
		tc := &tls.Config{
			ServerName: hc.Host,
		}
		tc1 := xoption.TLSConfig(rd)
		if tc1.InsecureSkipVerify {
			tc.InsecureSkipVerify = true
		}
		xoption.SetTLSConfig(opt, tc)
	}

	if opt.Empty() {
		return nil
	}
	return opt
}

var _ xrpc.Request = (*RequestNative)(nil)

type RequestNative struct {
	API     string
	Request *http.Request
}

func (h *RequestNative) Protocol() string {
	return "HTTP"
}

func (h *RequestNative) String() string {
	return "HTTPRequestNative:" + h.APIName()
}

func (h *RequestNative) APIName() string {
	if h.API != "" {
		return h.API
	}
	return h.Request.URL.Path
}

func (h *RequestNative) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
	req := h.Request.WithContext(ctx)
	if req.Host == HostDummy {
		req.Host = ""
	}
	setHTTPRequestUA(req)
	return req.Write(w)
}

func (h *RequestNative) balancer() xbalance.Reader {
	host := h.Request.URL.Hostname()
	if host == HostDummy {
		return nil
	}
	port := h.Request.URL.Port()
	if port == "" {
		if strings.EqualFold(h.Request.URL.Scheme, "http") {
			port = "80"
		} else if strings.EqualFold(h.Request.URL.Scheme, "https") {
			port = "443"
		}
	}
	return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", net.JoinHostPort(host, port)))
}

func (h *RequestNative) OptionReader(ctx context.Context, opt xoption.Reader) xoption.Reader {
	mp := xoption.NewMapOption()
	if ap := h.balancer(); ap != nil {
		xbalance.OptSetReader(mp, ap)
	}
	if tc := h.tslConfig(opt); tc != nil {
		xoption.SetTLSConfig(mp, tc)
	}
	return mp
}

func (h *RequestNative) tslConfig(rd xoption.Reader) *tls.Config {
	if !strings.EqualFold(h.Request.URL.Scheme, "https") {
		return nil
	}
	serverName := h.Request.Host
	tc := xoption.TLSConfig(rd)
	if tc != nil {
		if serverName != "" {
			tc = tc.Clone()
			tc.ServerName = serverName
		}
		return tc
	}
	return &tls.Config{
		ServerName: serverName,
	}
}

func setHTTPRequestUA(req *http.Request) {
	if req.UserAgent() == "" {
		req.Header.Set("User-Agent", "anygo-xrpc/1.0")
	}
}
