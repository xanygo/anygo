//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xhandler

import (
	"bytes"
	"maps"
	"net/http"
	"time"

	"github.com/xanygo/anygo/xcache"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xhttp"
)

// Cache 给 GET 请求添加缓存
type Cache struct {
	// Store 必填，缓存对象
	Store xcache.Cache[string, string]

	// Key 必填，缓存的 key，在 Handler 未执行前执行
	// 返回值的第一个参数是缓存的 key，第二个参数是缓存有效期，若为 0 则不缓存
	Key func(w http.ResponseWriter, r *http.Request) (string, time.Duration)
}

func (c *Cache) checkCan(w http.ResponseWriter, r *http.Request) bool {
	return r.Method == http.MethodGet && w.Header().Get("ETag") == ""

}

func (c *Cache) Next(handler http.Handler) http.Handler {
	cache := &xcache.TransString[*cachedResponse]{
		Cache: c.Store,
		Codec: xcodec.JSON,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.checkCan(w, r) {
			handler.ServeHTTP(w, r)
			return
		}
		key, ttl := c.Key(w, r)
		if key == "" || ttl <= 0 {
			handler.ServeHTTP(w, r)
			return
		}

		cr, err := cache.Get(r.Context(), key)
		if err == nil {
			cr.writeTo(w)
			return
		}

		bf := &bytes.Buffer{}
		wn := &bufferedResponseWriter{
			W:      w,
			Buffer: bf,
		}
		header1 := maps.Clone(w.Header())
		handler.ServeHTTP(wn, r)

		code := wn.GetStatusCode()
		if code == 0 || code == http.StatusOK {
			diff := xhttp.HeaderDiffMore(header1, w.Header())
			xhttp.WriteHeader(w, diff)
			cr := &cachedResponse{
				H: diff,
				B: bf.Bytes(),
			}
			_ = cache.Set(r.Context(), key, cr, ttl)
		}

		if code != 0 {
			w.WriteHeader(code)
		}
		if bf.Len() > 0 {
			_, _ = w.Write(bf.Bytes())
		}
	})
}

type cachedResponse struct {
	H http.Header
	B []byte
}

func (c *cachedResponse) writeTo(w http.ResponseWriter) {
	xhttp.WriteHeader(w, c.H)
	_, _ = w.Write(c.B)
}
