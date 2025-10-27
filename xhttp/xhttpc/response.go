//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
)

var _ xrpc.Response = (*Response)(nil)

type Response struct {
	Handler HandlerFunc

	resp    *http.Response
	readErr error
}

func (resp *Response) String() string {
	if resp.resp == nil {
		return "FetchResponse"
	}
	return "FetchResponse:" + resp.resp.Status
}

func (resp *Response) LoadFrom(ctx context.Context, req xrpc.Request, node *xnet.ConnNode, opt xoption.Reader) error {
	timeout := xoption.WriteTimeout(opt)
	if err := node.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer node.SetDeadline(time.Time{})

	maxSize := xoption.MaxResponseSize(opt)
	bio := bufio.NewReader(io.LimitReader(node, maxSize))
	resp.resp, resp.readErr = http.ReadResponse(bio, nil)
	if resp.readErr != nil {
		return resp.readErr
	}
	resp.readErr = resp.Handler(ctx, resp.resp)
	return resp.readErr
}

func (resp *Response) ErrCode() int64 {
	if resp.readErr != nil {
		return xerror.ErrCode(resp.readErr, 255)
	}
	if resp.resp != nil {
		return int64(resp.resp.StatusCode)
	}
	return 2
}

func (resp *Response) ErrMsg() string {
	if resp.readErr != nil {
		return resp.readErr.Error()
	}
	if resp.resp != nil {
		return resp.resp.Status
	}
	return "response not exists"
}

func (resp *Response) Response() *http.Response {
	return resp.resp
}

func (resp *Response) Unwrap() any {
	return resp.resp
}
