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
//	sepIndex: 拆分位置
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

// CutIndexBefore  将字符串从索引位置拆分为两部分,返回前面的部分
//
//	sepIndex: 拆分位置，支持超出字符串 s 正常的索引位置
//	当 index > len(s) 时，返回字符串整体
//	当 index <= 0     时，返回空字符串
func CutIndexBefore(s string, sepIndex int) (before string) {
	if sepIndex > len(s) {
		return s
	}
	if sepIndex <= 0 {
		return ""
	}
	return s[:sepIndex]
}

// CutIndexAfter  将字符串从索引位置拆分为两部分,返回后面的部分
//
//	sepIndex: 拆分位置，支持超出字符串 s 正常的索引位置
//	当 index > len(s) 时，返回空字符串
//	当 index <= 0     时，返回字符串整体
//	sepLen：拆分字符串的长度，应 >=0
func CutIndexAfter(s string, sepIndex int, sepLen int) (after string) {
	if sepIndex > len(s) {
		return ""
	}
	if sepIndex < 0 {
		return s
	}
	return substr(s[sepIndex:], sepLen)
}

// CutLastN 反向在字符串 s 中查找第 n 个子字符串 ,并拆分为前后两部分
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

// CutLastNAfter 反向在字符串 s 中查找第 n 个子字符串 ,并返回后面部分
func CutLastNAfter(s string, substr string, n int) (after string) {
	index := LastIndexN(s, substr, n)
	return CutIndexAfter(s, index, len(substr))
}

// CutLastByteN 反向在字符串 s 中查找第 n 个字符(c) ,并拆分为前后两部分
func CutLastByteN(s string, c byte, n int) (before string, after string, found bool) {
	index := LastIndexByteN(s, c, n)
	return CutIndex(s, index, 1)
}

// CutLastByteNBefore 反向在字符串 s 中查找第 n 个字符(c) ,并返回前面部分
func CutLastByteNBefore(s string, c byte, n int) (before string) {
	index := LastIndexByteN(s, c, n)
	return CutIndexBefore(s, index)
}

// CutLastByteNAfter 反向在字符串 s 中查找第 n 个 字符(c) ,并返回后面部分
func CutLastByteNAfter(s string, c byte, n int) (after string) {
	index := LastIndexByteN(s, c, n)
	return CutIndexAfter(s, index, 1)
}

// Substr 截取字符串
//
//	s: 待截取的字符串
//	start: 开始的位置，支持负数，0，正数索引位置，当为负数时，表示从字符尾部开始计数，
//	  如 -1 表示倒数第一个字符。
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

func HasPrefixFold(s string, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[0:len(prefix)], prefix)
}

func HasSuffixFold(s string, suffix string) bool {
	return len(s) >= len(suffix) && strings.EqualFold(s[len(s)-len(suffix):], suffix)
}

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
