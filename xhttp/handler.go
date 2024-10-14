//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-04

package xhttp

import (
	"encoding/json"
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

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
