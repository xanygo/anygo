//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zos"
	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/store/xkv/internal/file"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xio/xfs"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xpp"
)

var _ StringStorage = (*File)(nil)

func NewFile(dataDir string) *File {
	return &File{
		Dir: dataDir,
	}
}

// NewFileAny 创建一个值类型支持泛型类型的，使用文件系统存储的 KV 存储对象
func NewFileAny[V any](dataDir string, coder xcodec.Codec) *Transformer[V] {
	return &Transformer[V]{
		Codec:   coder,
		Storage: NewFile(dataDir),
	}
}

// File 基于本地文件系统的 KV 存储实现,值类型为 string
type File struct {
	// Dir 数据存储目录，必填
	Dir string

	// GC 触发清理多余空目录的间隔时间，可选
	// 若值 < 1秒，会使用默认值 300 秒
	GC time.Duration

	runner xpp.CooldownRunner

	groupMutex xsync.GroupMutex[any]
}

func (f *File) Init(param map[string]any) error {
	if f.Dir == "" {
		f.Dir, _ = xmap.GetString(param, "Dir")
		f.Dir = strings.TrimSpace(f.Dir)
		if f.Dir == "" {
			return errors.New("miss Dir")
		}
	}
	return nil
}

func (f *File) autoCompact() {
	f.runner.Run(f.GC, f.doCompact)
}

func (f *File) doCompact() {
	zos.GlobalLock()
	defer zos.GlobalUnlock()

	expire := time.Now().Add(-5 * time.Minute)
	deleted, err := xfs.RemoveEmptyDir(f.Dir, expire)
	if err != nil {
		xlog.Warn(context.Background(), "anygo_xkv_FileStorage_gc", xlog.Err("error", err))
	} else {
		xlog.Info(context.Background(), "anygo_xkv_FileStorage_gc", xlog.Int("deleted", deleted))
	}
}

func (f *File) Has(ctx context.Context, key string) (bool, error) {
	fb := &file.Base{
		Key:        key,
		Dir:        f.getDataDir(key),
		Type:       internal.DataTypeAny,
		GroupMutex: &f.groupMutex,
	}
	return fb.Has(ctx)
}

func (f *File) Delete(ctx context.Context, keys ...string) error {
	errs := make([]error, 0)
	for _, key := range keys {
		if err := f.deleteOne(key); err != nil {
			errs = append(errs, err)
		}
	}
	go f.autoCompact()
	return errors.Join(errs...)
}

func (f *File) deleteOne(key string) error {
	fp := f.getDataDir(key)
	err := os.RemoveAll(fp)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (f *File) getDataDir(key string) string {
	sg := md5.Sum([]byte("anygo" + key))
	s := hex.EncodeToString(sg[:])
	fp := filepath.Join(f.Dir, s[:2], s[2:4], s[4:6], s[6:])
	return fp
}

func (f *File) String(key string) String[string] {
	return &file.String{
		Base: &file.Base{
			Key:        key,
			Dir:        f.getDataDir(key),
			Type:       internal.DataTypeString,
			GroupMutex: &f.groupMutex,
		},
	}
}

func (f *File) List(key string) List[string] {
	return &file.List{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key:        key,
			Dir:        f.getDataDir(key),
			Type:       internal.DataTypeList,
			GroupMutex: &f.groupMutex,
		},
	}
}

func (f *File) Hash(key string) Hash[string] {
	return &file.Hash{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key:        key,
			Dir:        f.getDataDir(key),
			Type:       internal.DataTypeHash,
			GroupMutex: &f.groupMutex,
		},
	}
}

func (f *File) Set(key string) Set[string] {
	return &file.Set{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key:        key,
			Dir:        f.getDataDir(key),
			Type:       internal.DataTypeSet,
			GroupMutex: &f.groupMutex,
		},
	}
}

func (f *File) ZSet(key string) ZSet[string] {
	return &file.ZSet{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key:        key,
			Dir:        f.getDataDir(key),
			Type:       internal.DataTypeZSet,
			GroupMutex: &f.groupMutex,
		},
	}
}
