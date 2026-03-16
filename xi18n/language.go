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

type Language string

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
