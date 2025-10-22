//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-22

package xrpc

import (
	"context"
	"io"

	"github.com/xanygo/anygo/xoption"
)

// DiscardResponse 返回一个特殊的 response，
// 适用于 Request 已经处理完所有请求逻辑，并返回错误状态，不需要 Response 读取 Server 响应结果的情况
func DiscardResponse() Response {
	return discard
}

var discard = &discardResponse{}

var _ Response = (*discardResponse)(nil)

type discardResponse struct {
}

func (resp *discardResponse) String() string {
	return "discardResponse"
}

func (resp *discardResponse) LoadFrom(ctx context.Context, req Request, rd io.Reader, opt xoption.Reader) error {
	return nil
}

func (resp *discardResponse) ErrCode() int64 {
	return 0
}

func (resp *discardResponse) ErrMsg() string {
	return ""
}

func (resp *discardResponse) Unwrap() any {
	return nil
}
