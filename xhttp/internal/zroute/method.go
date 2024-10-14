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
