//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-08

package xfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// RemoveEmptyDir 查找并删除 ModTime 在 expire 之前的空目录
func RemoveEmptyDir(root string, expire time.Time) (int, error) {
	emptyDirs := make(map[string]bool, 10)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if path == root || err != nil {
			return err
		}
		parent := filepath.Dir(path)
		delete(emptyDirs, parent)
		if d.IsDir() {
			if expire.IsZero() {
				emptyDirs[path] = true
				return nil
			}

			info, err1 := d.Info()
			if err1 != nil {
				return err1
			}
			if info.ModTime().Before(expire) {
				emptyDirs[path] = true
			}

			return nil
		}
		return nil
	})
	if len(emptyDirs) == 0 {
		return 0, err
	}
	var deleted int
	var errs []error
	for dir := range emptyDirs {
		if err = os.RemoveAll(dir); err != nil {
			errs = append(errs, err)
		} else {
			deleted++
		}
	}
	if err != nil {
		errs = append(errs, err)
	}
	return deleted, errors.Join(errs...)
}
