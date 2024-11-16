//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-05

package xhttp

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/xanygo/anygo/xarchive"
	"github.com/xanygo/anygo/xslice"
)

type FSHandler interface {
	http.Handler
	Exists(fp string) bool
	Open(name string) (fs.File, error)
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

func (fs FSHandlers) Open(name string) (fs.File, error) {
	for _, f := range fs {
		if f.Exists(name) {
			return f.Open(name)
		}
	}
	return nil, os.ErrNotExist
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

func (e *FS) Open(name string) (fs.File, error) {
	return e.FS.Open(name)
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

func (z *ZipFile) Open(name string) (fs.File, error) {
	if er := z.Init(); er != nil {
		return nil, er
	}
	return z.initReader.Open(name)
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
	if fh, ok := h.(FSHandler); ok {
		return fh
	}
	return &fsWrap{Handler: h}
}

var _ FSHandler = (*fsWrap)(nil)

type fsWrap struct {
	http.Handler
}

func (f *fsWrap) Open(name string) (fs.File, error) {
	if hf, ok := f.Handler.(http.FileSystem); ok {
		return hf.Open(name)
	}
	return nil, fs.ErrNotExist
}

func (f *fsWrap) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	f.Handler.ServeHTTP(writer, request)
}

func (f *fsWrap) Exists(fp string) bool {
	if hf, ok := f.Handler.(http.FileSystem); ok {
		file, err := hf.Open(fp)
		if file != nil {
			_ = file.Close()
		}
		return err == nil
	}
	return true
}

var _ http.Handler = (*FSMerge)(nil)

type FSMerge struct {
	FS          FSHandler
	mergedFiles map[string]*mergeFile
}

func (fm *FSMerge) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filename := req.URL.Path
	if len(fm.mergedFiles) > 0 {
		file, ok := fm.mergedFiles[filename]
		if ok {
			if et := req.Header.Get("If-None-Match"); et != "" && et == file.Etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("Content-Type", file.ContentType)
			w.Header().Set("ETag", file.Etag)
			_, _ = w.Write(file.Body)
			return
		}
	}
	fm.FS.ServeHTTP(w, req)
}

func (fm *FSMerge) Merge(alias string, ct string, names ...string) error {
	if fm.mergedFiles == nil {
		fm.mergedFiles = make(map[string]*mergeFile)
	}
	mf, ok := fm.mergedFiles[alias]
	if ok {
		return fmt.Errorf("already has alias %q", alias)
	}
	mf = &mergeFile{
		ContentType: ct,
	}
	bf := &bytes.Buffer{}
	for _, name := range names {
		file, err := fm.FS.Open(name)
		if err != nil {
			return fmt.Errorf("merge %q failed：%w", name, err)
		}
		if _, err = bf.ReadFrom(file); err != nil {
			_ = file.Close()
			return fmt.Errorf("read %q failed: %w", name, err)
		}
		_ = file.Close()

		bf.WriteString("\n\n")
	}
	mf.Body = bf.Bytes()
	mf.Etag = fmt.Sprintf(`"%x"`, md5.Sum(bf.Bytes()))

	fm.mergedFiles[alias] = mf
	return nil
}

type mergeFile struct {
	ContentType string
	Body        []byte
	Etag        string
}