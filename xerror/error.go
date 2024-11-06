//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xerror

import "errors"

var (
	// NotFound 错误：数据找不到
	NotFound = NewCodeError(1000, "not found")

	InvalidStatus = NewCodeError(1001, "invalid status")
)

func IsNotFound(err error) bool {
	return errors.Is(err, NotFound)
}
