//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xhandler

import (
	"bytes"
	"net/http"
)

var _ http.ResponseWriter = (*bufferedResponseWriter)(nil)

type bufferedResponseWriter struct {
	statusCode int
	Buffer     *bytes.Buffer
	W          http.ResponseWriter
}

func (b *bufferedResponseWriter) Header() http.Header {
	return b.W.Header()
}

func (b *bufferedResponseWriter) Write(bytes []byte) (int, error) {
	return b.Buffer.Write(bytes)
}

func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

func (b *bufferedResponseWriter) GetStatusCode() int {
	if b.statusCode == 0 {
		return http.StatusOK
	}
	return b.statusCode
}
