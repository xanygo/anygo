//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-04

package xrpc

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xbalance"
	"github.com/xanygo/anygo/xnet/xservice"
	"github.com/xanygo/anygo/xoption"
)

var _ Request = (*HTTPRequest)(nil)

type HTTPRequest struct {
	API    string // APIName
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   io.Reader
}

func (r *HTTPRequest) String() string {
	return "HTTPRequest:" + r.APIName()
}

func (r *HTTPRequest) APIName() string {
	if r.API != "" {
		return r.API
	}
	return r.Path
}

func (r *HTTPRequest) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
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

func (r *HTTPRequest) getURL(so xoption.Reader, address string) (string, error) {
	opt := xservice.OptHTTP(so)
	var scheme string = "http"
	if opt.HTTPS {
		scheme = "https"
	}
	u, err := url.Parse(r.Path)
	if err != nil {
		return "", err
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

func (r *HTTPRequest) getMethod() string {
	if r.Method == "" {
		return http.MethodGet
	}
	return r.Method
}

func (r *HTTPRequest) OptionReader(ctx context.Context, rd xoption.Reader) xoption.Reader {
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

var _ Response = (*HTTPResponse)(nil)

type HTTPResponse struct {
	Handler func(ctx context.Context, resp *http.Response) error

	resp    *http.Response
	readErr error
}

func (resp *HTTPResponse) String() string {
	if resp.resp == nil {
		return "HTTPResponse"
	}
	return "HTTPResponse:" + resp.resp.Status
}

func (resp *HTTPResponse) LoadFrom(ctx context.Context, r io.Reader, opt xoption.Reader) error {
	bio := bufio.NewReader(r)
	resp.resp, resp.readErr = http.ReadResponse(bio, nil)
	if resp.readErr != nil {
		return resp.readErr
	}
	resp.readErr = resp.Handler(ctx, resp.resp)
	return resp.readErr
}

func (resp *HTTPResponse) ErrCode() int64 {
	if resp.readErr != nil {
		return xerror.ErrCode(resp.readErr, 255)
	}
	if resp.resp != nil {
		return int64(resp.resp.StatusCode)
	}
	return 2
}

func (resp *HTTPResponse) ErrMsg() string {
	if resp.readErr != nil {
		return resp.readErr.Error()
	}
	if resp.resp != nil {
		return resp.resp.Status
	}
	return "response not exists"
}

func (resp *HTTPResponse) Response() *http.Response {
	return resp.resp
}

var _ Request = (*HTTPRequestNative)(nil)

type HTTPRequestNative struct {
	API     string
	Request *http.Request
}

func (h *HTTPRequestNative) String() string {
	return "HTTPRequestNative:" + h.APIName()
}

func (h *HTTPRequestNative) APIName() string {
	if h.API != "" {
		return h.API
	}
	return h.Request.URL.Path
}

func (h *HTTPRequestNative) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
	req := h.Request.WithContext(ctx)
	setHTTPRequestUA(req)
	return req.Write(w)
}

func (h *HTTPRequestNative) balancer() xbalance.Reader {
	host := h.Request.URL.Hostname()
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

func (h *HTTPRequestNative) OptionReader(ctx context.Context, opt xoption.Reader) xoption.Reader {
	mp := xoption.NewMapOption()
	xbalance.OptSetReader(mp, h.balancer())
	if tc := h.tslConfig(opt); tc != nil {
		xoption.SetTLSConfig(mp, tc)
	}
	return mp
}

func (h *HTTPRequestNative) tslConfig(rd xoption.Reader) *tls.Config {
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
