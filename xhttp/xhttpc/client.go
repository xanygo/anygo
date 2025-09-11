//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-10

package xhttpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xnet/xrpc"
)

func Invoke(ctx context.Context, service string, req *http.Request, handler HandlerFunc, opts ...xrpc.Option) error {
	hr := &NativeRequest{
		Request: req,
	}
	resp := &Response{
		Handler: handler,
	}
	return xrpc.Invoke(ctx, service, hr, resp, opts...)
}

func Get(ctx context.Context, service string, url string, handler HandlerFunc, opts ...xrpc.Option) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	return Invoke(ctx, service, req, handler, opts...)
}

func InvokeWithCodec(ctx context.Context, service string, method string, url string, body any, ec xcodec.Encoder, handler HandlerFunc, opts ...xrpc.Option) error {
	var contentType string
	if hct, ok := ec.(xcodec.HasContentType); ok {
		contentType = hct.ContentType()
	} else {
		return errors.New("invalid codec: not xcodec.HasContentType")
	}

	bf, err := ec.Encode(body)
	if err != nil {
		return err
	}
	rd := bytes.NewBuffer(bf)
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	return Invoke(ctx, service, req, handler, opts...)
}

func Post(ctx context.Context, service string, url string, body io.Reader, ct string, handler HandlerFunc, opts ...xrpc.Option) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", ct)
	return Invoke(ctx, service, req, handler, opts...)
}

func PostForm(ctx context.Context, service string, url string, body url.Values, handler HandlerFunc, opts ...xrpc.Option) error {
	return InvokeWithCodec(ctx, service, http.MethodPost, url, body, xcodec.Form, handler, opts...)
}

func PostJSON(ctx context.Context, service string, url string, body any, handler HandlerFunc, opts ...xrpc.Option) error {
	return InvokeWithCodec(ctx, service, http.MethodPost, url, body, xcodec.JSON, handler, opts...)
}
