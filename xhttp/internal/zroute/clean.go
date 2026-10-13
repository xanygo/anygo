//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-12

package zroute

import (
	"bytes"
	"path"
	"strings"
)

// CleanPath 返回干净的 Path，去除掉 Path 中的 "/./","/../", 和连续的 "/"如 “//”
func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

// CleanPattern 归一化后的 pattern 地址,去掉变量的正则只保留变量名
func CleanPattern(p string) string {
	index := strings.IndexByte(p, ':')
	if index == -1 {
		return p
	}
	bf := &bytes.Buffer{}
	for {
		bf.WriteString(p[:index])
		end := strings.IndexByte(p[index:], '}')
		if end == -1 {
			panic("invalid pattern:" + p)
		}
		p = p[index+end:]
		index = strings.IndexByte(p, ':')
		if index == -1 {
			bf.WriteString(p)
			break
		}
	}
	return bf.String()
}
