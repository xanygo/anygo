//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xhandler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
)

// ETag 给所有 GET 请求，并且statusCode=200 的响应添加 etag 标记。
// 若请求携带 If-None-Match，并且和实际的 etag 一直，则直接发送  304 响应，不再发送 Body
type ETag struct {
	// Can 可选，用于判断当前请求是否需要添加 etag, 在业务 handler 执行前执行
	// 若为 nil，则跳过此判断
	Can func(w http.ResponseWriter, r *http.Request) bool
}

func (e *ETag) checkCan(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet ||
		w.Header().Get("ETag") != "" ||
		e.Can != nil && !e.Can(w, r) {
		return false
	}
	return true
}

func (e *ETag) Next(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !e.checkCan(w, r) {
			handler.ServeHTTP(w, r)
			return
		}
		bf := &bytes.Buffer{}
		wn := &bufferedResponseWriter{
			W:      w,
			Buffer: bf,
		}
		handler.ServeHTTP(wn, r)
		code := wn.GetStatusCode()

		if (code == 0 || code == http.StatusOK) && sendEtag(w, r, bf.Bytes()) {
			return
		}

		if code != 0 {
			w.WriteHeader(code)
		}
		if bf.Len() > 0 {
			_, _ = w.Write(bf.Bytes())
		}
	})
}

func sendEtag(w http.ResponseWriter, r *http.Request, bf []byte) bool {
	tag := getETag(bf)
	if tag != "" {
		w.Header().Set("ETag", tag)
		if match := r.Header.Get("If-None-Match"); match == tag {
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}
	return false
}

func getETag(bf []byte) string {
	if len(bf) == 0 {
		return ""
	}
	v := md5.Sum(bf)
	str := hex.EncodeToString(v[:])
	return fmt.Sprintf("%q", str)
}
