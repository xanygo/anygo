//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-18

package xstr

import "strings"

func HasAnyPrefix(str string, prefix ...string) bool {
	for _, v := range prefix {
		if strings.HasPrefix(str, v) {
			return true
		}
	}
	return false
}

func HasAnySuffix(str string, suffix ...string) bool {
	for _, v := range suffix {
		if strings.HasSuffix(str, v) {
			return true
		}
	}
	return false
}

// EqualAny 判断是否和任意一个字符串相等
func EqualAny(str string, values ...string) bool {
	for _, v := range values {
		if str == v {
			return true
		}
	}
	return false
}

func EqualFoldAny(str string, values ...string) bool {
	for _, v := range values {
		if strings.EqualFold(str, v) {
			return true
		}
	}
	return false
}
