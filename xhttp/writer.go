//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-04

package xhttp

import (
	"encoding/json"
	"net/http"
	"text/template"
)

// WriteJSON 输出 JSON 格式的数据，状态码为 200
func WriteJSON(w http.ResponseWriter, data any) {
	WriteJSONStatus(w, http.StatusOK, data)
}

// WriteJSONStatus 输出 JSON 格式的数据，需要指定状态码
func WriteJSONStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteText 输出 TEXT 格式的数据，状态码为 200
func WriteText(w http.ResponseWriter, text []byte) {
	WriteTextStatus(w, http.StatusOK, text)
}

// WriteTextStatus 输出 TEXT 格式的数据，需要指定状态码
func WriteTextStatus(w http.ResponseWriter, status int, text []byte) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(text)
}

// WriteHTML 输出 HTML 格式的数据，状态码为 200
func WriteHTML(w http.ResponseWriter, html []byte) {
	WriteHTMLStatus(w, http.StatusOK, html)
}

// WriteHTMLStatus 输出 HTML 格式的数据，需要指定状态码
func WriteHTMLStatus(w http.ResponseWriter, status int, html []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(html)
}

func WriteJSONP(w http.ResponseWriter, cb string, data any) {
	WriteJSONPStatus(w, http.StatusOK, cb, data)
}

func WriteJSONPStatus(w http.ResponseWriter, status int, cb string, data any) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(status)
	if cb == "" {
		_, _ = w.Write([]byte("callback("))
	} else {
		template.JSEscape(w, []byte(cb))
		_, _ = w.Write([]byte("("))
	}
	_ = json.NewEncoder(w).Encode(data)
	_, _ = w.Write([]byte(")"))
}

var _ http.ResponseWriter = (*StatusWriter)(nil)

type StatusWriter struct {
	W          http.ResponseWriter
	statusCode int
	wrote      int
}

func (w *StatusWriter) Header() http.Header {
	return w.W.Header()
}

func (w *StatusWriter) Write(bytes []byte) (int, error) {
	n, err := w.W.Write(bytes)
	w.wrote += n
	return n, err
}

func (w *StatusWriter) WriteHeader(statusCode int) {
	w.W.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *StatusWriter) Wrote() int {
	return w.wrote
}

func (w *StatusWriter) StatusCode() int {
	return w.statusCode
}

func (w *StatusWriter) Unwrap() http.ResponseWriter {
	return w.W
}
