//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-22

package xrpc

import (
	"context"
	"io"

	"github.com/xanygo/anygo/ds/xoption"
)

// NoRequest 返回一个不会发送请求信息的 Request 对象
func NoRequest() Request {
	return noReq
}

var noReq = &noRequest{}

var _ Request = (*noRequest)(nil)

type noRequest struct{}

func (n noRequest) String() string {
	return "noRequest"
}

func (n noRequest) Protocol() string {
	return "empty"
}

func (n noRequest) APIName() string {
	return "empty"
}

func (n noRequest) WriteTo(ctx context.Context, w io.Writer, opt xoption.Reader) error {
	return nil
}

// NoResponse 返回一个特殊的 response，
// 适用于 Request 已经处理完所有请求逻辑，并返回错误状态，不需要 Response 读取 Server 响应结果的情况
func NoResponse() Response {
	return noResp
}

var noResp = &noResponse{}

var _ Response = (*noResponse)(nil)

type noResponse struct {
}

func (resp *noResponse) String() string {
	return "noResponse"
}

func (resp *noResponse) LoadFrom(ctx context.Context, req Request, r io.Reader, opt xoption.Reader) error {
	return nil
}

func (resp *noResponse) ErrCode() int64 {
	return 0
}

func (resp *noResponse) ErrMsg() string {
	return ""
}

func (resp *noResponse) Unwrap() any {
	return nil
}
