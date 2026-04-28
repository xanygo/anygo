//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-08

package xi18n

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/xanygo/anygo/ds/xctx"
	"github.com/xanygo/anygo/ds/xslice"
)

// HTTPHandler  读取 HTTP 的 Accept-Language 和 cookie 中存储的首选项信息的中间件
type HTTPHandler struct {
	// LangCookieName cookie 中存储首选语言的字段名，可选，当为空时默认值为 lang
	LangCookieName string

	// Allowed 从 cookie 中读取的首选语言的有效值，可选，当不为空时生效
	// 若为空，则允许 Bundle 里所有语言
	Allowed []Language

	// Bundle 可选,语言资源
	Bundle *Bundle

	// LanguageResolver 可选，从请求中返回首选语言列表，
	// 若为 nil，或者返回结果为空，则默认从 Header 的 Accept-Language 读取
	LanguageResolver func(req *http.Request) []Language
}

func (h *HTTPHandler) getCookieName() string {
	if h.LangCookieName == "" {
		return "lang"
	}
	return h.LangCookieName
}

var ctxKeyHTTPHandler = xctx.NewKey()

func (h *HTTPHandler) Next(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := h.acceptLanguages(r)
		ctx := context.WithValue(r.Context(), ctxKeyHTTPHandler, h)
		if len(accept) > 0 {
			ctx = ContextWithLanguages(ctx, accept)
		}
		if h.Bundle != nil {
			ctx = ContextWithBundle(ctx, h.Bundle, "")
		}
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}

func (h *HTTPHandler) PreferredLanguage(req *http.Request) []Language {
	if oh, ok := req.Context().Value(ctxKeyHTTPHandler).(*HTTPHandler); ok {
		return oh.acceptLanguages(req)
	}
	return h.acceptLanguages(req)
}

func (h *HTTPHandler) acceptLanguages(req *http.Request) []Language {
	var accept []Language
	if h.LanguageResolver != nil {
		accept = h.LanguageResolver(req)
	}
	if len(accept) == 0 {
		accept = ParserAccept(req.Header.Get("Accept-Language"))
	}
	// 读取以设置到 cookie 中的首选语言
	if ck, err := req.Cookie(h.getCookieName()); err == nil && len(ck.Value) > 0 {
		cv := Language(ck.Value)
		if len(h.Allowed) == 0 || xslice.ContainsAny(h.Allowed, cv) {
			accept = slices.Insert(accept, 0, cv)
		}
	}
	return accept
}

// DomainResolver 按照域名前缀，和 Bundle 里的 language 对比，解析出首选语言。
// 如域名是 zh.example.com, Bundle 里有 zh 语言，则返回首选语言为 zh。
//
// 该方法可赋值给 LanguageResolver 属性作为回调函数
func (h *HTTPHandler) DomainResolver(req *http.Request) []Language {
	host := req.Host
	if host == "" || h.Bundle == nil {
		return nil
	}
	for _, lang := range h.Bundle.Languages() {
		prefix := string(lang) + "."
		if strings.HasPrefix(host, prefix) {
			return []Language{lang}
		}
	}
	return nil
}
