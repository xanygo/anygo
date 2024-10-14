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

func WriteJSON(w http.ResponseWriter, data any) {
	WriteJSONStatus(w, http.StatusOK, data)
}

func WriteJSONStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteText(w http.ResponseWriter, text []byte) {
	WriteTextStatus(w, http.StatusOK, text)
}

func WriteTextStatus(w http.ResponseWriter, status int, text []byte) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(text)
}

func WriteHTML(w http.ResponseWriter, html []byte) {
	WriteHTMLStatus(w, http.StatusOK, html)
}

func WriteHTMLStatus(w http.ResponseWriter, status int, html []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(html)
}
