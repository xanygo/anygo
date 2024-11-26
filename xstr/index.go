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

// LastIndexByteN 反向在字符串 s 中查找第 n 个 字符(c) 的位置，若查找不到会返回 -1
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

// BytePairIndex 查找字符串中，前后匹配的字符对，
// 可以来查找一对匹配的:<>、()、{}，支持内部嵌套，
// 比如对于字符串 “(hello(a,b,c,d(e,f))) word(a,b)”，查找 '(' 和 ')'，
// 会找到 "(hello(a,b,c,d(e,f)))" 的首个 '(' 的索引位置和最后一个 ‘）’的索引位置
//
// 返回值：
//
//	leftIndex: 首个 left 的索引位置
//	rightIndex: 当 ok=true 时，值为最后一个 right 的索引位置；当 ok=false 时，值为最后读取到的 right 的索引位置
//	ok: 是否正确的匹配
func BytePairIndex(str string, left byte, right byte) (leftIndex int, rightIndex int, ok bool) {
	leftIndex = -1
	rightIndex = -1
	var count int
	for index := 0; index < len(str); index++ {
		switch str[index] {
		case left:
			count++
			if !ok {
				ok = true
				leftIndex = index
			}
		case right:
			count--
			if count == 0 {
				return leftIndex, index, true
			}
			if count < 0 { // right 在 left 之前出现，如 ") hello ("
				return leftIndex, index, false
			}
			rightIndex = index
		}
	}
	return leftIndex, rightIndex, false
}
