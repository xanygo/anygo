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

// ParserAccept 解析 HTTP Header 中的 Accept-Language 字段
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

// ContextWithLanguages 将当前应该使用的语言信息设置到 context 里。
// 若是 languages 是多个，则优先级支持的语言排在最前面
func ContextWithLanguages(ctx context.Context, languages []Language) context.Context {
	return context.WithValue(ctx, ctxKeyLang, languages)
}

// LanguagesFromContext 从 ctx 里读取应该使用的语言列表.
// 优先级支持的语言排在最前面。
// 如返回 []Language { "zh", "en"},表明 优先使用语言 zh（中文），其次才是 en (英文)
func LanguagesFromContext(ctx context.Context) []Language {
	result, _ := ctx.Value(ctxKeyLang).([]Language)
	return result
}
