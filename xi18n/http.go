//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-08

package xi18n

import (
	"net/http"
	"slices"

	"github.com/xanygo/anygo/ds/xslice"
)

// HTTPLanguageHandler  读取 HTTP 的 Accept-Language 和 cookie 中存储的首选项信息的中间件
type HTTPLanguageHandler struct {
	// CookieName cookie 中存储首选语言的字段名，可选，当为空时默认值为 lang
	CookieName string

	// Allow 从 cookie 中读取的首选语言的有效值，可选，当不为空时生效
	Allow []Language

	// Bundle 可选
	Bundle *Bundle
}

func (h HTTPLanguageHandler) getCookieName() string {
	if h.CookieName == "" {
		return "lang"
	}
	return h.CookieName
}

func (h HTTPLanguageHandler) Next(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := ParserAccept(r.Header.Get("Accept-Language"))

		// 读取以设置到 cookie 中的首选语言
		if ck, err := r.Cookie(h.getCookieName()); err == nil && len(ck.Value) > 0 {
			cv := Language(ck.Value)
			if len(h.Allow) == 0 || xslice.ContainsAny(h.Allow, cv) {
				accept = slices.Insert(accept, 0, Language(ck.Value))
			}
		}
		ctx := r.Context()
		if len(accept) > 0 {
			ctx = ContextWithLanguages(r.Context(), accept)
		}
		if h.Bundle != nil {
			ctx = ContextWithBundle(ctx, h.Bundle, "")
		}
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}
