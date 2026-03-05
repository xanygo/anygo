//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-04

package xfs

import (
	"errors"
	"io/fs"
)

var _ fs.FS = (OverlayFS)(nil)

// OverlayFS 是一个“叠加文件系统”，用于将多个 fs.FS 合并成一个统一的虚拟文件系统。
//
// 核心功能：
// OverlayFS 按顺序维护一组底层文件系统（fs.FS）。当访问文件或目录时，
// 会按照顺序遍历底层 FS，返回第一个命中的文件或目录，实现“前置覆盖、后置默认”的效果。
//
// 优先级规则：
// 1. 文件优先：如果多个 FS 中存在同名文件，优先返回最前面的 FS。
// 2. 目录合并：同名目录会自动合并，ReadDir 会返回所有子 FS 中的唯一条目。
// 3. 文件与目录冲突：
//   - 如果前面的 FS 是文件，后面的目录会被覆盖。
//   - 如果前面的 FS 是目录，后面的目录会合并。
type OverlayFS []fs.FS

func (of OverlayFS) Open(name string) (fs.File, error) {
	for _, fsys := range of {
		if f, err := fsys.Open(name); err == nil {
			return f, nil
		}
	}

	return nil, fs.ErrNotExist
}

var _ fs.StatFS = (OverlayFS)(nil)

func (of OverlayFS) Stat(name string) (fs.FileInfo, error) {
	for _, fsys := range of {
		info, err := fs.Stat(fsys, name)
		if err == nil {
			return info, nil
		}
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return nil, err
	}
	return nil, fs.ErrNotExist
}

var _ fs.ReadDirFS = (OverlayFS)(nil)

func (of OverlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	seen := make(map[string]fs.DirEntry)
	for _, fsys := range of {
		entries, err := fs.ReadDir(fsys, name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if _, exists := seen[e.Name()]; !exists {
				seen[e.Name()] = e
			}
		}
	}
	if len(seen) == 0 {
		return nil, fs.ErrNotExist
	}

	result := make([]fs.DirEntry, 0, len(seen))
	for _, e := range seen {
		result = append(result, e)
	}
	return result, nil
}
