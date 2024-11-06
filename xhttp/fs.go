//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-05

package xhttp

import (
	"io/fs"
	"net/http"
	"path"
)

var _ http.Handler = (*FSHandler)(nil)

type FSHandler struct {
	// FS 必填，可以是 embed.FS
	FS fs.FS

	// RootDir 可选，FS 里此目录作为基准目录
	RootDir string

	etag etagStore
}

func (e *FSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var fileName string
	if e.RootDir != "" {
		fileName = path.Join(e.RootDir, r.URL.Path)
	} else {
		fileName = r.URL.Path
	}
	if e.etag.hasSameETag(w, r, e.FS, fileName) {
		return
	}
	http.ServeFileFS(w, r, e.FS, fileName)
}
