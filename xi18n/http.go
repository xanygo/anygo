//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-08

package xi18n

import (
	"net/http"
	"slices"

	"github.com/xanygo/anygo/xslice"
)

type HTTPHandlerLanguage struct {
	CookieName string
	Handler    http.Handler
	Allow      []Language
}

func (h HTTPHandlerLanguage) getCookieName() string {
	if h.CookieName == "" {
		return "lang"
	}
	return h.CookieName
}

func (h HTTPHandlerLanguage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := ParserAccept(r.Header.Get("Accept-Language"))

	// 读取以设置到 cookie 中的首选语言
	if ck, err := r.Cookie(h.getCookieName()); err == nil && len(ck.Value) > 0 {
		cv := Language(ck.Value)
		if len(h.Allow) == 0 || xslice.ContainsAny(h.Allow, cv) {
			accept = slices.Insert(accept, 0, Language(ck.Value))
		}
	}

	if len(accept) > 0 {
		r = r.WithContext(ContextWithLanguages(r.Context(), accept))
	}
	h.Handler.ServeHTTP(w, r)
}
