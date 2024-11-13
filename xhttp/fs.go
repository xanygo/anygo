//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-05

package xhttp

import (
	"io/fs"
	"net/http"
	"path"
	"sync"
)

var _ http.Handler = (*FSHandler)(nil)

// FSHandler  将 fs.FS 文件封装为 http.Handler，以使其内部文件可被访问到
// 相比  http.FS 而言，可指定 fs 内部的基准目录，并添加了 etag 支持
type FSHandler struct {
	// FS 必填，可以是 embed.FS
	FS fs.FS

	// RootDir 可选，FS 里此目录作为基准目录
	RootDir string

	// NotFound 可选，当文件不存在时的回调
	NotFount http.Handler

	etag      etagStore
	fileNames map[string]bool
	once      sync.Once
}

func (e *FSHandler) init() {
	e.fileNames = make(map[string]bool)
	root := "."
	if e.RootDir != "" {
		root = e.RootDir
	}
	_ = fs.WalkDir(e.FS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		e.fileNames[path] = true
		return nil
	})
}

func (e *FSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hasNotFound := e.NotFount != nil
	if hasNotFound {
		e.once.Do(e.init)
	}
	var fileName string
	if e.RootDir != "" {
		fileName = path.Join(e.RootDir, r.URL.Path)
	} else {
		fileName = r.URL.Path
	}
	if e.etag.hasSameETag(w, r, e.FS, fileName) {
		return
	}
	if hasNotFound && !e.fileNames[fileName] {
		e.NotFount.ServeHTTP(w, r)
		return
	}
	http.ServeFileFS(w, r, e.FS, fileName)
}

// Exists 判断文件是否存在，fp 应该时经过了 path.Clean 的，并且不以 / 开头
func (e *FSHandler) Exists(fp string) bool {
	e.once.Do(e.init)
	return e.fileNames[fp]
}
