//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import "context"

type ctxKey uint8

const (
	ctxKeyLang ctxKey = iota
)

func ContextWithLanguages(ctx context.Context, languages []Language) context.Context {
	return context.WithValue(ctx, ctxKeyLang, languages)
}

func LanguagesFromContext(ctx context.Context) []Language {
	result, _ := ctx.Value(ctxKeyLang).([]Language)
	return result
}
