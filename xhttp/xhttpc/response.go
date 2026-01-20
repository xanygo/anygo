//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/ds/xsync"
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
	resp.resp = nil
	resp.readErr = resp.doLoadFrom(ctx, req, node, opt)
	if resp.readErr == nil {
		return nil
	}

	// 包裹错误，让 rpc client 的 retryPolicy 可以依据 error 来判断是否能重试
	// 只有特定的请求 Method 和 StatusCode 才允许重试
	// 如 GET 请求，响应为 500，则标记为临时错误，允许重试
	var te xerror.TemporaryFailure
	if !errors.As(resp.readErr, &te) {
		temp := retryableStatus(resp.resp.StatusCode)
		return xerror.WithTemporary(resp.readErr, temp)
	}
	return resp.readErr
}

func retryableStatus(code int) bool {
	switch code {
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (resp *Response) doLoadFrom(ctx context.Context, req xrpc.Request, node *xnet.ConnNode, opt xoption.Reader) error {
	timeout := xoption.ReadTimeout(opt)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := node.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer node.SetDeadline(time.Time{})

	maxSize := xoption.MaxResponseSize(opt)
	bio := bufio.NewReader(io.LimitReader(node, maxSize))
	rr, err := http.ReadResponse(bio, nil)
	if err != nil {
		return fmt.Errorf("http.ReadResponse %w", err)
	}
	resp.decompress(rr)
	resp.resp = rr
	err = resp.Handler(ctx, rr)
	if err == nil {
		return nil
	}
	return fmt.Errorf("resp.Handler %w", err)
}

func (resp *Response) decompress(rr *http.Response) {
	if strings.EqualFold(rr.Header.Get("Content-Encoding"), "gzip") {
		rr.Body = &gzipReader{body: rr.Body}
		rr.Header.Del("Content-Encoding")
		rr.Header.Del("Content-Length")
		rr.ContentLength = -1
		rr.Uncompressed = true
	}
}

type gzipReader struct {
	body      io.ReadCloser
	zr        *gzip.Reader
	zerr      error
	closeOnce xsync.OnceDoErr
}

var errReadOnClosedResBody = errors.New("xhttpc: read on closed response body")

func (gz *gzipReader) Read(p []byte) (n int, err error) {
	if gz.closeOnce.Done() {
		return 0, errReadOnClosedResBody
	}
	if gz.zr == nil {
		if gz.zerr == nil {
			gz.zr, gz.zerr = gzip.NewReader(gz.body)
		}
		if gz.zerr != nil {
			return 0, gz.zerr
		}
	}

	if err != nil {
		return 0, err
	}
	return gz.zr.Read(p)
}

func (gz *gzipReader) Close() error {
	return gz.closeOnce.Do(func() error {
		return gz.body.Close()
	})
}

func (resp *Response) ErrCode() int64 {
	if resp.readErr != nil {
		return xerror.ErrCode(resp.readErr, 500)
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
