//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package zhttp

import (
	_ "embed"
	"net/http"
	"path"
	"time"

	"github.com/xanygo/anygo/xattr"
)

//go:embed asset/404.webp
var image404 []byte

var assetLastMTime time.Time

func init() {
	var err error
	assetLastMTime, err = time.Parse(time.DateTime, "2024-10-01 10:00:00")
	if err != nil {
		panic(err)
	}
}

// HandlerImage404 处理图片类请求的 404
func HandlerImage404(w http.ResponseWriter, r *http.Request) bool {
	ext := path.Ext(r.URL.Path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".tif":
	default:
		return false
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Expires", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	if ims := r.Header.Get("If-Modified-Since"); ims != "" {
		ifModifiedSince, err := time.Parse(http.TimeFormat, ims)
		if err == nil && !assetLastMTime.After(ifModifiedSince) {
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}

	w.Header().Set("Last-Modified", assetLastMTime.UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(image404)
	return true
}

func HandlerStatic404(w http.ResponseWriter, r *http.Request) bool {
	ext := path.Ext(r.URL.Path)
	switch ext {
	case ".css", ".js":
		w.WriteHeader(http.StatusNotFound)
		if xattr.RunMode() == xattr.ModeDebug {
			w.Write([]byte(r.URL.Path + "\n"))
		}
		w.Write([]byte("404 not found"))
		return true
	default:
		return false
	}
}
