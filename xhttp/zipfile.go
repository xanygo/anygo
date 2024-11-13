//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-05

package xhttp

import (
	"archive/zip"
	"bytes"
	"errors"
	"net/http"
	"path"
	"sync"

	"github.com/xanygo/anygo/xarchive"
	"github.com/xanygo/anygo/xslice"
)

var _ http.Handler = (*ZipFileHandler)(nil)

// ZipFileHandler 将 zip 文件封装为 http.Handler，以使其内部文件可被访问到，并添加了 etag 支持
type ZipFileHandler struct {
	// Reader 压缩文件的 reader， Reader 和 Content 二者至少有一个，并优先使用 Reader
	Reader *zip.Reader

	// Content zip 压缩的内容，可选
	Content []byte

	// RootDir 可选，FS 里此目录作为基准目录
	RootDir string

	// NotFound 可选，当文件不存在时的回调
	NotFound http.Handler

	etag etagStore

	initReader    *zip.Reader
	initErr       error
	initFileNames map[string]bool
	once          sync.Once
}

func (z *ZipFileHandler) init() {
	if z.Reader != nil {
		z.initReader = z.Reader
		z.initFileNames = xslice.ToMap(xarchive.ZipFileNames(z.Reader, 0), true)
		return
	}
	if len(z.Content) == 0 {
		z.initErr = errors.New("both Reader and Content are empty")
		return
	}
	z.initReader, z.initErr = zip.NewReader(bytes.NewReader(z.Content), int64(len(z.Content)))
	if z.initReader != nil {
		z.initFileNames = xslice.ToMap(xarchive.ZipFileNames(z.initReader, 0), true)
	}
}

func (z *ZipFileHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	z.once.Do(z.init)
	if z.initErr != nil {
		http.Error(w, z.initErr.Error(), http.StatusInternalServerError)
		return
	}
	fileName := req.URL.Path
	if z.RootDir != "" {
		fileName = path.Join(z.RootDir, fileName)
	}

	if z.etag.hasSameETag(w, req, z.initReader, fileName) {
		return
	}
	if z.NotFound != nil && !z.initFileNames[fileName] {
		z.NotFound.ServeHTTP(w, req)
		return
	}
	http.ServeFileFS(w, req, z.initReader, fileName)
}

// Exists 判断文件是否存在，fp 应该时经过了 path.Clean 的，并且不以 / 开头
func (z *ZipFileHandler) Exists(fp string) bool {
	z.once.Do(z.init)
	return len(z.initFileNames) > 0 && z.initFileNames[fp]
}
