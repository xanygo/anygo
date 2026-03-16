//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import "context"

type ctxKey uint8

const (
	ctxKeyLang ctxKey = iota
	ctxKeyBundle
)

type ctxBundle struct {
	bundle    *Bundle
	namespace string
}

// ContextWithBundle 将本地化资源信息存储到 ctx 里去，如此之后可以直接在 .go 文件中使用 RA 和 RA 等系列函数渲染文本内容
func ContextWithBundle(ctx context.Context, b *Bundle, namespace string) context.Context {
	rr := &ctxBundle{
		bundle:    b,
		namespace: namespace,
	}
	return context.WithValue(ctx, ctxKeyBundle, rr)
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
