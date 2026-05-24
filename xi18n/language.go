//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-06

package xi18n

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

// Language 语言的类型，如 zh、zh-CN、en、en-US 等
type Language string

// SameFamily 判断语言是否是同语言族
//
//	如 “zh”，“zh-CN”，"zh-TW" 都是同语言族
func (l Language) SameFamily(lang Language) bool {
	if l == lang {
		return true
	}
	return l.Base() == lang.Base()
}

// Base 返回语言的基础语言标签（primary language subtag）。
//
// 例如：
//
//	zh           -> zh
//	zh-CN        -> zh
//	zh-Hans-CN   -> zh
//	en-US        -> en
//
// 该方法仅提取第一个 '-' 之前的部分，常用于判断两个语言是否属于同一语言族。
func (l Language) Base() Language {
	if p, _, ok := strings.Cut(string(l), "-"); ok {
		return Language(p)
	}
	return l
}

const (
	LangZh   Language = "zh"    // 中文
	LangZhCN Language = "zh-CN" // 中文-简体
	LangZhHK Language = "zh-HK" // 中文-香港
	LangZhTW Language = "zh-TW" // 中文-台湾

	LangEn   Language = "en"    // 英文
	LangEnUS Language = "en-US" // 英文-美国
	LangEnGB Language = "en-GB" // 英文-英国
)

type languageWithQ struct {
	name Language
	q    float64
}

// ParserAccept 解析 HTTP Header 中的 Accept-Language 字段
//
// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
func ParserAccept(accept string) []Language {
	arr := strings.Split(accept, ",")
	result := make([]languageWithQ, 0, len(arr))
	for _, v := range arr {
		v = strings.TrimSpace(v)
		b, q, _ := strings.Cut(v, ";")
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		var qf = 1.0
		if strings.HasPrefix(q, "q=") {
			if ff, err := strconv.ParseFloat(q[2:], 64); err == nil {
				qf = ff
			}
		}
		if qf <= 0 {
			continue
		}
		result = append(result, languageWithQ{
			name: Language(b),
			q:    qf,
		})
	}
	slices.SortFunc(result, func(a, b languageWithQ) int {
		return cmp.Compare(b.q, a.q)
	})

	ret := make([]Language, 0, len(result))
	for _, v := range result {
		ret = append(ret, v.name)
	}
	return ret
}
