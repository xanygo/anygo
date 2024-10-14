//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-13

package xerror

import (
	"encoding"
	"encoding/json"
	"errors"
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

type HasErrCode interface {
	ErrCode() int64
}

func NewCodeError(err error, code int64) *CodeError {
	return &CodeError{
		Code: code,
		Err:  err,
	}
}

var _ encoding.TextMarshaler = (*CodeError)(nil)
var _ json.Marshaler = (*CodeError)(nil)
var _ HasErrCode = (*CodeError)(nil)

type CodeError struct {
	Err  error
	Code int64 // 错误码
	Data any   // 导致错误的数据，可选
}

func (c *CodeError) Error() string {
	if c.Err == nil {
		return "<nil>"
	}
	return c.Err.Error()
}

func (c *CodeError) ErrCode() int64 {
	return c.Code
}

func (c *CodeError) ErrData() any {
	return c.Data
}

func (c *CodeError) Unwrap() error {
	return c.Err
}

func (c *CodeError) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"Code": c.Code,
		"Msg":  c.Error(),
		"Data": c.Data,
	}
	return json.Marshal(data)
}

func (c *CodeError) MarshalText() (text []byte, err error) {
	return c.MarshalJSON()
}
