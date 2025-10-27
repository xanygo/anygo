//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xfs

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// HasFd 有实现 Fd 方法
type HasFd interface {
	Fd() uintptr
}

// Exists 判断文件/目录是否存在
func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// KeepDirExists 保持文件夹存在，若不存在则创建
// 若路径为文件，则删除，然后创建文件夹
func KeepDirExists(dir string) error {
	info, err := os.Stat(dir)
	if err == nil && info.IsDir() {
		return nil
	}
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		err1 := os.MkdirAll(dir, 0777)
		if err1 == nil || errors.Is(err1, fs.ErrExist) {
			return nil
		}
		return err1
	}
	if err != nil {
		return err
	}

	// 若不是目录，则删除掉
	if err = os.Remove(dir); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	err = os.MkdirAll(dir, 0777)
	if err == nil || errors.Is(err, fs.ErrExist) {
		return nil
	}
	return err
}

// CleanFiles 按照文件前缀清理文件
//
//	pattern: eg /home/work/logs/access_log.log.*
//	remaining: 文件保留个数，eq ：24
func CleanFiles(pattern string, remaining int) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(files) <= remaining {
		return nil
	}

	type finfo struct {
		info os.FileInfo
		path string
	}

	var infos []*finfo
	for _, p := range files {
		info, err := os.Stat(p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			log.Fatalf("[anygo][cleanFiles] os.Stat(%q) has error:%v\n", p, err)
			continue
		}
		infos = append(infos, &finfo{path: p, info: info})
	}

	if len(infos) <= remaining {
		return nil
	}

	sort.Slice(infos, func(i, j int) bool {
		a := Ctime(infos[i].info)
		b := Ctime(infos[j].info)
		return b.Before(a)
	})

	for i := remaining; i < len(infos); i++ {
		p := infos[i].path
		if err = os.Remove(p); err != nil {
			log.Printf("[anygo][cleanFiles] os.Remove(%q), err=%v\n", p, err)
		}
	}
	return nil
}
