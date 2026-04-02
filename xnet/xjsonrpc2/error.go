//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-28

package xjsonrpc2

import (
	"encoding/json"

	"github.com/xanygo/anygo/xerror"
)

var _ error = (*Error)(nil)
var _ xerror.CodeError = (*Error)(nil)

type Error struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e Error) ErrCode() int64 {
	return e.Code
}

func (e Error) Error() string {
	return e.Message
}

var (
	ErrParse          = &Error{Code: -32700, Message: "parse error"}     // 数据包格式错误
	ErrInvalidRequest = &Error{Code: -32600, Message: "invalid request"} // 请求
	ErrMethodNotFound = &Error{Code: -32601, Message: "method not found"}
	ErrInvalidParams  = &Error{Code: -32602, Message: "invalid params"} // 解析请求中的参数失败
	ErrInternal       = &Error{Code: -32603, Message: "internal error"}

	ErrInvalidResult = &Error{Code: -32000, Message: "invalid result"} // 解析响应的 结果失败
)
