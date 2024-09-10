//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-06

package xi18n

import (
	"context"
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

func ParserAccept(accept string) []Language {
	arr := strings.Split(accept, ",")
	result := make([]Language, 0, len(arr))
	for _, v := range arr {
		v = strings.TrimSpace(v)
		b, _, _ := strings.Cut(v, ";")
		b = strings.TrimSpace(b)
		if b != "" {
			result = append(result, Language(b))
		}
	}
	return result
}

func ContextWithLanguages(ctx context.Context, languages []Language) context.Context {
	return context.WithValue(ctx, ctxKeyLang, languages)
}

func LanguagesFromContext(ctx context.Context) []Language {
	result, _ := ctx.Value(ctxKeyLang).([]Language)
	return result
}
