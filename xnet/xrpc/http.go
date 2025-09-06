//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-04

package xrpc

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet/xservice"
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

func (r *HTTPRequest) WriteTo(ctx context.Context, conn net.Conn, opt *xservice.Option) error {
	api, err := r.getURL(opt, conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, r.getMethod(), api, r.Body)
	if err != nil {
		return err
	}
	return req.Write(conn)
}

func (r *HTTPRequest) getURL(so *xservice.Option, address string) (string, error) {
	opt := so.HTTP
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
	readErr error
}

func (resp *HTTPResponse) String() string {
	// TODO implement me
	panic("implement me")
}

func (resp *HTTPResponse) ReadFrom(r io.Reader) (int64, error) {
	bio := bufio.NewReader(r)
	resp.resp, resp.readErr = http.ReadResponse(bio, nil)
	var readLen int64
	if resp.resp != nil {
		defer resp.resp.Body.Close()
		body, err := io.ReadAll(resp.resp.Body)
		readLen = int64(len(body))
		if err != nil {
			resp.readErr = err
			return readLen, err
		}
		resp.resp.Body = io.NopCloser(bytes.NewReader(body))
	}
	return readLen, resp.readErr
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
