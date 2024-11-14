//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-05

package xhttp

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"sync"

	"github.com/xanygo/anygo/xarchive"
	"github.com/xanygo/anygo/xslice"
)

type FSHandler interface {
	http.Handler
	Exists(fp string) bool
}

type FSHandlers []FSHandler

func (fs FSHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(fs) == 0 {
		http.NotFound(w, r)
		return
	}
	for i := 0; i < len(fs)-1; i++ {
		if fs[i].Exists(r.URL.Path) {
			fs[i].ServeHTTP(w, r)
			return
		}
	}
	fs[len(fs)-1].ServeHTTP(w, r)
}

func (fs FSHandlers) Exists(fp string) bool {
	for _, f := range fs {
		if f.Exists(fp) {
			return true
		}
	}
	return false
}

var _ FSHandler = (*FS)(nil)

// FS  将 fs.FS 文件封装为 http.Handler，以使其内部文件可被访问到
// 相比  http.FS 而言，可指定 fs 内部的基准目录，并添加了 etag 支持
type FS struct {
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

func (e *FS) init() {
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

func (e *FS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.once.Do(e.init)
	fileName := fullFSPath(e.RootDir, r.URL.Path)
	if e.etag.hasSameETag(w, r, e.FS, fileName) {
		return
	}
	if e.fileNames[fileName] {
		http.ServeFileFS(w, r, e.FS, fileName)
		return
	}
	if e.NotFount != nil {
		e.NotFount.ServeHTTP(w, r)
		return
	}
	NotFoundHandler().ServeHTTP(w, r)
}

// Exists 判断文件是否存在，fp 应该时经过了 path.Clean 的，并且不以 / 开头
func (e *FS) Exists(fp string) bool {
	e.once.Do(e.init)
	fileName := fullFSPath(e.RootDir, fp)
	return e.fileNames[fileName]
}

func fullFSPath(root string, fp string) string {
	if root == "" {
		return fp
	}
	return path.Join(root, fp)
}

var _ FSHandler = (*ZipFile)(nil)

// ZipFile 将 zip 文件封装为 http.Handler，以使其内部文件可被访问到，并添加了 etag 支持
type ZipFile struct {
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

func (z *ZipFile) Init() error {
	z.once.Do(z.init)
	return z.initErr
}

func (z *ZipFile) init() {
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

func (z *ZipFile) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if z.initErr != nil {
		http.Error(w, z.initErr.Error(), http.StatusInternalServerError)
		return
	}
	fileName := fullFSPath(z.RootDir, req.URL.Path)

	if z.etag.hasSameETag(w, req, z.initReader, fileName) {
		return
	}
	if z.initFileNames[fileName] {
		http.ServeFileFS(w, req, z.initReader, fileName)
		return
	}
	if z.NotFound != nil {
		z.NotFound.ServeHTTP(w, req)
		return
	}
	NotFoundHandler().ServeHTTP(w, req)
}

// Exists 判断文件是否存在，fp 应该时经过了 path.Clean 的，并且不以 / 开头
func (z *ZipFile) Exists(fp string) bool {
	z.once.Do(z.init)
	return len(z.initFileNames) > 0 && z.initFileNames[fullFSPath(z.RootDir, fp)]
}

func ToFSHandler(h http.Handler) FSHandler {
	return &fsWrap{Handler: h}
}

var _ FSHandler = (*fsWrap)(nil)

type fsWrap struct {
	http.Handler
}

func (f *fsWrap) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	f.Handler.ServeHTTP(writer, request)
}

func (f *fsWrap) Exists(fp string) bool {
	return true
}
