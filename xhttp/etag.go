//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-05

package xhttp

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sync"
)

type etagStore struct {
	values sync.Map
}

func (z *etagStore) getETag(fsSystem fs.FS, fp string) string {
	if val, ok := z.values.Load(fp); ok {
		return val.(string)
	}
	file, err := fsSystem.Open(fp)
	if err != nil {
		return ""
	}
	defer file.Close()
	m := md5.New()
	if _, err = io.Copy(m, file); err != nil {
		return ""
	}
	str := hex.EncodeToString(m.Sum(nil))
	result := fmt.Sprintf("%q", str)
	z.values.Store(fp, result)
	return result
}

func (z *etagStore) hasSameETag(w http.ResponseWriter, req *http.Request, rd fs.FS, fileName string) bool {
	tag := z.getETag(rd, fileName)
	if tag == "" {
		return false
	}
	w.Header().Set("ETag", tag)
	if match := req.Header.Get("If-None-Match"); match == tag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}
