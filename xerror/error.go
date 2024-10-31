//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xerror

import "errors"

var NotFound = NewCodeError(1404, "Not Found")

func IsNotFound(err error) bool {
	return errors.Is(err, NotFound)
}
