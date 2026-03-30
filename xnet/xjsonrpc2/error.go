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
	ErrParse          = &Error{Code: -32700, Message: "parse error"}
	ErrInvalidRequest = &Error{Code: -32600, Message: "invalid request"}
	ErrMethodNotFound = &Error{Code: -32601, Message: "method not found"}
	ErrInvalidParams  = &Error{Code: -32602, Message: "invalid params"}
	ErrInternal       = &Error{Code: -32603, Message: "internal error"}
)
