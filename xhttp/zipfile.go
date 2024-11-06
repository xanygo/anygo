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

	"github.com/xanygo/anygo/xsync"
)

var _ http.Handler = (*ZipFileHandler)(nil)

type ZipFileHandler struct {
	// Reader 压缩文件的 reader， Reader 和 Content 二者至少有一个，并优先使用 Reader
	Reader *zip.Reader

	// Content zip 压缩的内容，可选
	Content []byte

	// RootDir 可选，FS 里此目录作为基准目录
	RootDir string

	etag etagStore
	once xsync.OnceDoValue2[*zip.Reader, error]
}

func (z *ZipFileHandler) initReader() (*zip.Reader, error) {
	if z.Reader != nil {
		return z.Reader, nil
	}
	if len(z.Content) == 0 {
		return nil, errors.New("both Reader and Content are empty")
	}
	return zip.NewReader(bytes.NewReader(z.Content), int64(len(z.Content)))
}

func (z *ZipFileHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rd, err := z.once.Do(z.initReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileName := req.URL.Path
	if z.RootDir != "" {
		fileName = path.Join(z.RootDir, fileName)
	}

	if z.etag.hasSameETag(w, req, rd, fileName) {
		return
	}
	http.ServeFileFS(w, req, rd, fileName)
}
