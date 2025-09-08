//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-04

package xrpc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

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

var _ Response = (*HTTPResponse)(nil)

type HTTPResponse struct {
	resp    *http.Response
	Handler func(ctx context.Context, resp *http.Response) error
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
	return req.Write(w)
}

var _ xbalance.HasReader = (*HTTPRequestNative)(nil)

func (h *HTTPRequestNative) Balancer() xbalance.Reader {
	host := h.Request.URL.Hostname()
	port := h.Request.URL.Port()
	return xbalance.NewStaticByAddr(xnet.NewAddr("tcp", net.JoinHostPort(host, port)))
}
