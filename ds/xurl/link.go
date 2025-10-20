//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-27

package xurl

import (
	"net/url"

	"github.com/xanygo/anygo/internal/zbase"
)

// BaseLink 基于当前已有 url，附加额外参数，生成新连接。
// param 的参数可以覆盖 base 里的 query。
// 如 param=“q=你好&page=1&name=”，其中的值为空的情况， “name=” 将删除 base 中 的 name query 参数
func BaseLink(base *url.URL, queryPair ...any) string {
	if len(queryPair) == 0 {
		return base.String()
	}
	if len(queryPair)%2 != 0 {
		panic("queryPair length must be even")
	}
	query := base.Query()
	for i := 0; i < len(queryPair); i += 2 {
		key := queryPair[i].(string)
		value := zbase.ToString(queryPair[i+1])
		if value == "" {
			query.Del(key)
		} else {
			query.Set(key, value)
		}
	}
	if len(query) == 0 {
		return base.Path
	}
	return base.Path + "?" + query.Encode()
}

// NewLink 基于当前 url 的 path，生成新的 url，会丢掉 base 的所有 query 参数
func NewLink(base *url.URL, queryPair ...any) string {
	if len(queryPair) == 0 {
		return base.String()
	}
	query := url.Values{}
	for i := 0; i < len(queryPair); i += 2 {
		key := queryPair[i].(string)
		value := zbase.ToString(queryPair[i+1])
		if value != "" {
			query.Set(key, value)
		}
	}
	if len(query) == 0 {
		return base.Path
	}
	return base.Path + "?" + query.Encode()
}
