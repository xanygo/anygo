//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr

import (
	"math"
	"strings"
)

// CutIndex 将字符串从索引位置拆分为两部分
//
//	sepIndex: 拆分位置, 应 >= 0,此字符会被删除掉，CutIndex("abc",1,1) -> "a","c",true
//	sepLen：拆分字符串的长度，应 >= 0
func CutIndex(s string, sepIndex int, sepLen int) (before, after string, ok bool) {
	if sepIndex < 0 || sepIndex > len(s) {
		return s, "", false
	}
	return s[:sepIndex], substr(s[sepIndex:], sepLen), true
}

func substr(str string, index int) string {
	if index > len(str) {
		return ""
	}
	return str[index:]
}

// CutIndexBefore  将字符串从索引位置拆分为两部分,返回 index 前面的部分，相当于更安全的 str[:index]，
// 即使 index 超出长度也不会 panic
//
//	index: 拆分位置，支持超出字符串 s 正常的索引位置
//	当 index >= len(s) 时，返回字符串整体
//	当 index <= 0     时，返回空字符串
func CutIndexBefore(s string, index int) (before string) {
	if index >= len(s) {
		return s
	}
	if index <= 0 {
		return ""
	}
	return s[:index]
}

// CutIndexAfter  将字符串从索引位置拆分为两部分,返回 index 后面的部分，相当于更安全的 str[index+1:],
// 即使 index 超出长度也不会 panic。返回的内容不包含 index 本身
//
//	index: 拆分位置，支持超出字符串 s 正常的索引位置
//	当 index >= len(s) 时，返回空字符串
//	当 index < 0     时，返回字符串整体
func CutIndexAfter(s string, index int) (after string) {
	if index >= len(s) {
		return ""
	}
	if index < 0 {
		return s
	}
	return s[index+1:]
}

// CutLastN 反向在字符串 s 中查找第 n ( n >=0 ) 个子字符串 ,并拆分为前后两部分
//
// n 从 0 开始计数，即操作首个是  n=0
func CutLastN(s string, substr string, n int) (before string, after string, found bool) {
	index := LastIndexN(s, substr, n)
	if index < 0 {
		return s, "", false
	}
	return CutIndex(s, index, len(substr))
}

// CutLastNBefore 反向在字符串 s 中查找第 n 个子字符串 ,并返回前面部分
func CutLastNBefore(s string, substr string, n int) (before string) {
	index := LastIndexN(s, substr, n)
	return CutIndexBefore(s, index)
}

// CutLastNAfter 反向在字符串 s 中查找第 n ( n>=0 ) 个子字符串 ,并返回后面部分
func CutLastNAfter(s string, substr string, n int) (after string) {
	index := LastIndexN(s, substr, n)
	return CutIndexAfter(s, index)
}

// CutLastByteN 反向在字符串 s 中查找第 n ( n>=0 ) 个字符(c) ,并拆分为前后两部分
func CutLastByteN(s string, c byte, n int) (before string, after string, found bool) {
	index := LastIndexByteN(s, c, n)
	return CutIndex(s, index, 1)
}

// CutLastByteNBefore 反向在字符串 s 中查找第 n ( n >= 0 ) 个字符(c) ,并返回前面部分
func CutLastByteNBefore(s string, c byte, n int) (before string) {
	index := LastIndexByteN(s, c, n)
	return CutIndexBefore(s, index)
}

// CutLastByteNAfter 反向在字符串 s 中查找第 n 个 字符(c) ,并返回后面部分
func CutLastByteNAfter(s string, c byte, n int) (after string) {
	index := LastIndexByteN(s, c, n)
	return CutIndexAfter(s, index)
}

// Substr 安全的，截取字符串，相当于 str[start:end], 但是即使超出索引不会 panic
//
//	s: 待截取的字符串
//	start: 开始的位置，支持负数，0，正数索引位置，当为负数时，表示索引位置从字符尾部开始计数。
//	  如 -1 表示从倒数第一个字符开始，往后截取 length 长度的字符串，
//	length： 截取长度，允许超过字符串 s 的最大长度,应 > 0。若 <=0 则返回空。
func Substr(s string, start, length int) string {
	if s == "" || length <= 0 {
		return ""
	}
	if start < 0 {
		start += len(s)
		if start < 0 {
			start = 0
		}
	}
	end := start + length
	if end > len(s) {
		return s[start:]
	}
	return s[start:end]
}

// HasPrefixFold 检查字符串是否有不区分大小写相同的前缀
func HasPrefixFold(s string, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[0:len(prefix)], prefix)
}

// HasSuffixFold 检查字符串是否有不区分大小写相同的后缀
func HasSuffixFold(s string, suffix string) bool {
	return len(s) >= len(suffix) && strings.EqualFold(s[len(s)-len(suffix):], suffix)
}

// SplitLen 按照长度将字符串拆分为子串
func SplitLen(s string, length int) []string {
	total := math.Ceil(float64(len(s)) / float64(length))
	result := make([]string, 0, int(total))
	for i := 0; i < len(s); i += length {
		end := i + length
		if end > len(s) {
			end = len(s)
		}
		result = append(result, s[i:end])
	}
	return result
}
