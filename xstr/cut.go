//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

// CutIndex 将字符串从索引位置拆分为两部分
//
//	index: 拆分位置，支持超出字符串 s 正常的索引位置
func CutIndex(s string, index int) (before, after string) {
	if index > len(s) {
		return s, ""
	}
	if index <= 0 {
		return "", s
	}
	return s[:index], s[index:]
}
