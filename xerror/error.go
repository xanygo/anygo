//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xerror

import (
	"errors"
	"io/fs"
	"strconv"
	"strings"
)

type TraceError interface {
	error
	TraceData() map[string]any
}

const (
	CodeNotFound = iota + 1000
	CodeAlreadyExist
	CodeInvalidParam
	CodeDuplicateKey
	CodeClosed
)

var (
	NotFound     = NewCodeError(CodeNotFound, "not found")                 // 错误：数据找不到
	Closed       = NewCodeError(CodeClosed, "closed")                      // 错误：已关闭
	AlreadyExist = NewCodeError(CodeAlreadyExist, "already exists")        // 错误：已存在
	InvalidParam = NewCodeError(CodeInvalidParam, "invalid param")         // 错误：无效的请求参数
	DuplicateKey = NewCodeError(CodeDuplicateKey, "duplicate primary key") // 错误：重复的主键
)

// IsNotFound 判断是否资源不存在错误
func IsNotFound(err error) bool {
	if errors.Is(err, NotFound) || errors.Is(err, fs.ErrNotExist) {
		return true
	}
	var ae NotExistsError
	if errors.As(err, &ae) {
		return ae.NotExists()
	}
	txt := err.Error()
	// 其他的情况，比如 gorm.ErrRecordNotFound
	return strings.Contains(txt, "not found") || strings.Contains(txt, "does not exist")
}

// IsAlreadyExists 判断是否已存在错误
func IsAlreadyExists(err error) bool {
	if errors.Is(err, AlreadyExist) || errors.Is(err, fs.ErrExist) {
		return true
	}
	txt := err.Error()
	// 其他的情况，比如 TopK: key already exists
	return strings.Contains(txt, "already exists")
}

type NotExistsError interface {
	NotExists() bool
}

// IsInvalidParam 判断是否参数不对的错误
func IsInvalidParam(err error) bool {
	return errors.Is(err, InvalidParam)
}

func String(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func NewStatusError(code int64) *StatusError {
	return &StatusError{
		Code: code,
	}
}

var _ CodeError = (*StatusError)(nil)

// StatusError 状态异常的错误，常用语 rpc response 的校验
type StatusError struct {
	Code int64
}

func (s *StatusError) Error() string {
	return "invalid status " + strconv.FormatInt(s.Code, 10)
}

func (s *StatusError) ErrCode() int64 {
	return s.Code
}
