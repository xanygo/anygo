//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-13

package xerror

import (
	"encoding"
	"encoding/json"
	"errors"
	"strconv"
)

// ErrCode2 尝试读取错误码，若 err == nil，则总是返回 0, true。
// 若 error 实现了 HasErrCode 接口，则读取成功，否则失败。
func ErrCode2(err error) (code int64, ok bool) {
	if err == nil {
		return 0, true
	}
	var ec HasErrCode
	if errors.As(err, &ec) {
		return ec.ErrCode(), true
	}
	return 0, false
}

// ErrCode 读取错误码，若读取失败，则返回默认值 def
// 若 err==nil，总是返回 0
func ErrCode(err error, def int64) int64 {
	code, ok := ErrCode2(err)
	if ok {
		return code
	}
	return def
}

// CodeError 带有错误码的 error 接口定义
type CodeError interface {
	error
	HasErrCode
}

type (
	HasErrCode interface {
		ErrCode() int64
	}

	HasErrMsg interface {
		ErrMsg() string
	}
)

var _ error = (*codeError1)(nil)
var _ HasErrCode = (*codeError1)(nil)

func NewCodeError(code int64, msg string) CodeError {
	return &codeError1{
		Code: code,
		Msg:  msg,
	}
}

func fmtCode(code int64) string {
	return " (errno:" + strconv.FormatInt(code, 10) + ") "
}

type codeError1 struct {
	Code int64
	Msg  string
}

func (e *codeError1) ErrCode() int64 {
	return e.Code
}

func (e *codeError1) Error() string {
	return e.Msg + fmtCode(e.Code)
}

func WithCode(err error, code int64) CodeError {
	return &codeError2{
		Code: code,
		Err:  err,
	}
}

var _ encoding.TextMarshaler = (*codeError2)(nil)
var _ json.Marshaler = (*codeError2)(nil)
var _ CodeError = (*codeError2)(nil)

type codeError2 struct {
	Err  error
	Code int64 // 错误码
}

func (c *codeError2) Error() string {
	if c.Err == nil {
		return "<nil>"
	}
	return c.Err.Error() + fmtCode(c.Code)
}

func (c *codeError2) ErrCode() int64 {
	return c.Code
}

func (c *codeError2) Unwrap() error {
	return c.Err
}

func (c *codeError2) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"Code": c.Code,
		"Msg":  c.Error(),
	}
	return json.Marshal(data)
}

func (c *codeError2) MarshalText() (text []byte, err error) {
	return c.MarshalJSON()
}
