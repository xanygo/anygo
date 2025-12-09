//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xerror

import (
	"errors"
	"strings"
)

type TraceError interface {
	error
	TraceData() map[string]any
}

const (
	CodeNotFound = iota + 1000
	CodeInvalidStatus
	CodeInvalidParam
	CodeDuplicateKey
	CodeClosed
)

var (
	// NotFound 错误：数据找不到
	NotFound = NewCodeError(CodeNotFound, "not found")
	Closed   = NewCodeError(CodeClosed, "closed")

	// InvalidStatus 错误：数据的状态非正常
	InvalidStatus = NewCodeError(CodeInvalidStatus, "invalid status")

	// InvalidParam 错误：无效的请求参数
	InvalidParam = NewCodeError(CodeInvalidParam, "invalid param")
	DuplicateKey = NewCodeError(CodeDuplicateKey, "duplicate primary key")
)

// IsNotFound 判断是否资源不存在错误
func IsNotFound(err error) bool {
	if errors.Is(err, NotFound) {
		return true
	}
	var ae NotExistsError
	if errors.As(err, &ae) {
		return ae.NotExists()
	}
	txt := err.Error()
	// 其他的情况，比如 gorm.ErrRecordNotFound
	return strings.Contains(txt, "not found")
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
