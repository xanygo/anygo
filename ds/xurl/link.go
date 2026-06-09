//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-27

package xurl

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

// WithQuery 基于当前已有 url，附加额外参数，生成新连接。
//
// queryPair 的参数可以覆盖 base 里的 query。
// 如 "q" "你好" "page" "1" "name" "" 相当于 "q=你好&page=1&name=", 其中的值为空的情况， “name=” 将删除 base 中 的 name query 参数
func WithQuery(base *url.URL, queryPair ...string) (string, error) {
	if len(queryPair) == 0 {
		return base.String(), nil
	}
	if len(queryPair)%2 != 0 {
		return "", errors.New("queryPair length must be even")
	}
	query := base.Query()
	for i := 0; i < len(queryPair); i += 2 {
		key := queryPair[i]
		value := queryPair[i+1]
		if value == "" {
			query.Del(key)
		} else {
			query.Set(key, value)
		}
	}
	if len(query) == 0 {
		return base.Path, nil
	}
	return base.Path + "?" + query.Encode(), nil
}

// WithNewQuery 基于当前 url 的 path，生成新的 url，会丢掉 base 的所有 query 参数
func WithNewQuery(base *url.URL, queryPair ...string) (string, error) {
	if len(queryPair) == 0 {
		return base.String(), nil
	}
	if len(queryPair)%2 != 0 {
		return "", errors.New("queryPair length must be even")
	}
	query := url.Values{}
	for i := 0; i < len(queryPair); i += 2 {
		key := queryPair[i]
		value := queryPair[i+1]
		if value != "" {
			query.Set(key, value)
		}
	}
	if len(query) == 0 {
		return base.Path, nil
	}
	return base.Path + "?" + query.Encode(), nil
}

// PathJoin 连接 url 地址
//
//	http://example.com/hello/ + /world.html       ==> http://example.com/world.html
//	http://example.com/hello/ + world.html        ==> http://example.com/hello/world.html
//	https://example.com/hello/ + world.html?q=1   ==> http://example.com/hello/world.html?q=1
//
// base: 必须是一个有效的 url 地址
// rel:  相对地址，如 /world.html 或者 world.html，若以 "/" 开头，则表示是根目录。可以包含 query 参数
func PathJoin(base string, rel string) (string, error) {
	bu, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	bu.Fragment = ""

	before, qs, found := strings.Cut(rel, "?")
	bu.RawQuery = qs
	if found {
		rel = before
	}

	if strings.HasPrefix(rel, "/") {
		bu.Path = rel
		return bu.String(), nil
	}
	if !strings.HasSuffix(bu.Path, "/") {
		bu.Path = path.Dir(bu.Path)
	}
	bu.Path = path.Join(bu.Path, rel)
	return bu.String(), nil
}
