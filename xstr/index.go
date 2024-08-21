//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

import "strings"

// IndexN 在字符串 s 中查找第 N 个 substr 的位置，若查找不到会返回 -1
func IndexN(s string, substr string, n int) int {
	var index int
	for i := 0; i < n; i++ {
		pos := strings.Index(s, substr)
		if pos == -1 {
			return -1
		}
		if i < n-1 {
			z := pos + len(substr)
			index += z
			s = s[z:]
		} else {
			index += pos
		}
	}
	return index
}

// LastIndexN 反向在字符串 s 中查找第 N 个 substr 的位置，若查找不到会返回 -1
func LastIndexN(s string, substr string, n int) int {
	var pos int
	for i := 0; i < n; i++ {
		pos = strings.LastIndex(s, substr)
		if pos == -1 {
			return -1
		}
		if i < n-1 {
			s = s[:pos]
		}
	}
	return pos
}

// IndexByteN 在字符串 s 中查找第 N 个 byte 的位置，若查找不到会返回 -1
func IndexByteN(s string, c byte, n int) int {
	var index int
	for i := 0; i < n; i++ {
		pos := strings.IndexByte(s, c)
		if pos == -1 {
			return -1
		}
		if i < n-1 {
			z := pos + 1
			index += z
			s = s[z:]
		} else {
			index += pos
		}
	}
	return index
}

// LastIndexByteN 反向在字符串 s 中查找第 N 个 byte 的位置，若查找不到会返回 -1
func LastIndexByteN(s string, c byte, n int) int {
	var pos int
	for i := 0; i < n; i++ {
		pos = strings.LastIndexByte(s, c)
		if pos == -1 {
			return -1
		}
		if i < n-1 {
			s = s[:pos]
		}
	}
	return pos
}
