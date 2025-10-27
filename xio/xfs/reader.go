//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-27

package xfs

import (
	"io/fs"
	"os"
)

// CachedReader 若文件没有变化，返回之前已读取的数据
type CachedReader struct {
	Path string

	lastStats fs.FileInfo
	content   []byte
}

func (cf *CachedReader) Reset() {
	cf.lastStats = nil
}

func (cf *CachedReader) ReadFile() (content []byte, cache bool, err error) {
	info, err := os.Stat(cf.Path)
	if err != nil {
		return nil, false, err
	}
	if os.SameFile(cf.lastStats, info) && info.ModTime().Equal(cf.lastStats.ModTime()) {
		return cf.content, true, nil
	}
	content, err = os.ReadFile(cf.Path)
	if err == nil {
		cf.content = content
		cf.lastStats = info
	}
	return content, false, err
}
