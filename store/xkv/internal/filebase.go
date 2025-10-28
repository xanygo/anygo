//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package internal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/xio/xfs"
)

const kvDataFileExt = ".kvd"

type FileMeta struct {
	Key     string   `json:"k"`
	Type    DataType `json:"t"`
	Updated int64    `json:"c"`
}

type FileBase struct {
	Key string
	Dir string
}

func (fb FileBase) getMetaFilePath() string {
	return filepath.Join(fb.Dir, "meta")
}

func (fb FileBase) MetaFileStats() (os.FileInfo, error) {
	return os.Stat(fb.getMetaFilePath())
}

func (fb FileBase) SaveMeta(tp DataType) error {
	fp := fb.getMetaFilePath()
	old, _ := fb.loadMeta()
	if old.Type > DataTypeUnset && old.Type != tp {
		// 若数据类型不匹配，则清空原有的数据
		err := os.RemoveAll(fb.Dir)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	if err := xfs.KeepDirExists(fb.Dir); err != nil {
		return err
	}
	meta := FileMeta{
		Key:     fb.Key,
		Updated: time.Now().Unix(),
		Type:    tp,
	}
	bf, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(fp, bf, 0666)
}

func (fb FileBase) loadMeta() (FileMeta, error) {
	bf, err := os.ReadFile(fb.getMetaFilePath())
	if err != nil {
		return FileMeta{}, err
	}
	meta := &FileMeta{}
	if err := json.Unmarshal(bf, meta); err != nil {
		return FileMeta{}, err
	}
	return *meta, nil
}

// WriteKVDataFile 写 kv 数据文件
func (fb FileBase) WriteKVDataFile(baseName string, data string) (err error) {
	return fb.WriteFile(baseName+kvDataFileExt, data)
}

func (fb FileBase) WriteKVDataFile2(baseName string, data string) (added bool, err error) {
	baseName = baseName + kvDataFileExt
	fp := filepath.Join(fb.Dir, baseName)
	dir := filepath.Dir(fp)
	if err = xfs.KeepDirExists(dir); err != nil {
		return false, err
	}
	info, err := os.Stat(fp)
	// 判断此文件是否新增
	added = info == nil && errors.Is(err, fs.ErrNotExist)
	return added, os.WriteFile(fp, []byte(data), 0666)
}

func (fb FileBase) WriteFile(baseName string, data string) error {
	fp := filepath.Join(fb.Dir, baseName)
	dir := filepath.Dir(fp)
	if err := xfs.KeepDirExists(dir); err != nil {
		return err
	}
	return os.WriteFile(fp, []byte(data), 0666)
}

// CheckReadKVDataFile 检查并读取 kv 文件，
func (fb FileBase) CheckReadKVDataFile(baseName string, typ DataType, delete bool) (string, bool, error) {
	meta, err := fb.loadMeta()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if meta.Type != typ {
		return "", false, ErrInvalidType
	}

	return fb.ReadFile(baseName+kvDataFileExt, delete)
}

func (fb FileBase) ReadFile(baseName string, delete bool) (string, bool, error) {
	fp := filepath.Join(fb.Dir, baseName)
	bf, err := os.ReadFile(fp)
	if delete {
		_ = os.Remove(fp)
	}
	if err == nil {
		return unsafe.String(unsafe.SliceData(bf), len(bf)), true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	return "", false, err
}

func (fb FileBase) DeleteKVDataFile(baseName string) error {
	return fb.DeleteFile(baseName + kvDataFileExt)
}

func (fb FileBase) DeleteFile(baseName string) error {
	fp := filepath.Join(fb.Dir, baseName)
	return fb.OsRemove(fp)
}

func (fb FileBase) OsRemove(fp string) error {
	err := os.Remove(fp)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (fb FileBase) Md5(field string) string {
	sm := md5.Sum([]byte("anygo" + field))
	return hex.EncodeToString(sm[:])
}

func (fb FileBase) RangeKVFiles(ctx context.Context, typ DataType, fn func(path string, d fs.DirEntry) error) error {
	meta, err := fb.loadMeta()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if meta.Type != typ {
		return ErrInvalidType
	}
	err = fs.WalkDir(os.DirFS(fb.Dir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}

		if !strings.HasSuffix(d.Name(), kvDataFileExt) {
			return nil
		}
		err = fn(path, d)
		if err != nil && errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	})

	return err
}
