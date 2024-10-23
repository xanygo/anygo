//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-12

package zroute

import (
	"bytes"
	"path"
	"strings"
)

func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.CleanPath removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
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
