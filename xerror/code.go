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

func ErrCode(err error) (int64, bool) {
	var ec HasErrCode
	if errors.As(err, &ec) {
		return ec.ErrCode(), true
	}
	return 0, false
}

func ErrCode2(err error, def int64) int64 {
	code, ok := ErrCode(err)
	if ok {
		return code
	}
	return def
}

type CodeError interface {
	error
	HasErrCode
}

type HasErrCode interface {
	ErrCode() int64
}

var _ error = (*codeError1)(nil)
var _ HasErrCode = (*codeError1)(nil)

func NewCodeError(code int64, msg string) CodeError {
	return &codeError1{
		Code: code,
		Msg:  msg,
	}
}

func fmtCode(code int64) string {
	return "[code=" + strconv.FormatInt(code, 10) + "] "
}

type codeError1 struct {
	Code int64
	Msg  string
}

func (e *codeError1) ErrCode() int64 {
	return e.Code
}

func (e *codeError1) Error() string {
	return fmtCode(e.Code) + e.Msg
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
	return fmtCode(c.Code) + c.Err.Error()
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
