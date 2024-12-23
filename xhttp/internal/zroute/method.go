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

// GetPrefixMethod 获取字符串的 Method 前缀
// 如 GetUser -> GET，GetUserList ->  GET
// index  ->  GET
// PostXXX  -> POST
// DeleteByID -> DELETE
// Save ->  POST  // 以 Save 为前缀的都返回 POST
func GetPrefixMethod(s string) string {
	var index int
	for ; index < len(s); index++ {
		char := s[index]
		if index > 0 && char >= 'A' && char <= 'Z' {
			break
		}
	}
	method := s[:index]

	switch method {
	case "Index", "Search":
		return http.MethodGet
	case "Save", "Upload":
		return http.MethodPost
	case "Update":
		return http.MethodPut
	}

	if IsMethod(method) {
		return strings.ToUpper(method)
	}

	return http.MethodGet
}
