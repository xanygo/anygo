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

var (
	// NotFound 错误：数据找不到
	NotFound = NewCodeError(1000, "not found")

	// InvalidStatus 错误：数据的状态非正常
	InvalidStatus = NewCodeError(1001, "invalid status")

	// InvalidParam 错误：无效的请求参数
	InvalidParam = NewCodeError(1002, "invalid param")
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
