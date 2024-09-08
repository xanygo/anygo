//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-08

package xi18n

import "net/http"

func HandlerAcceptLanguage(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := ParserAccept(r.Header.Get("Accept-Language"))
		if len(accept) > 0 {
			r = r.WithContext(ContextWithLanguages(r.Context(), accept))
		}
		h.ServeHTTP(w, r)
	})
}
