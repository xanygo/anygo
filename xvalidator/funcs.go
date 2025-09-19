//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-02

package xvalidator

import (
	"fmt"
	"strings"
)

// IsHTTPURL 是否有效的 HTTP URL 地址
func IsHTTPURL(str string) error {
	scheme, _, ok := strings.Cut(str, "://")
	if !ok || (scheme != "http" && scheme != "https") {
		return fmt.Errorf("%q is not HTTP url", str)
	}
	return nil
}

func StringIn(name, value string, values ...string) error {
	for _, v := range values {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("%s=%q is not in %q", name, value, values)
}
