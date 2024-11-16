//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-13

package zroute

import (
	"net/http"
	"strings"
)

func CleanMethod(method string) string {
	return strings.ToUpper(method)
}

var allMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}
var methods = map[string]string{}

func init() {
	for _, method := range allMethods {
		methods[method] = method
	}
}

func IsMethod(method string) bool {
	method = strings.ToUpper(method)
	_, ok := methods[method]
	return ok
}

// SplitCamelCase 将驼峰函数名拆分为 2 部分，
// 如 GetUser -> Get,User ; GetUserList ->  Get, UserList
func SplitCamelCase(s string) (string, string) {
	if len(s) == 0 {
		return "", ""
	}

	var splitIndex int
	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			splitIndex = i
			break
		}
	}

	if splitIndex == 0 {
		return s, ""
	}
	return s[:splitIndex], s[splitIndex:]
}
