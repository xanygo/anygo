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
	"time"

	"github.com/xanygo/anygo/internal/zos"
	"github.com/xanygo/anygo/store/xkv/internal/file"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xio/xfs"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xpp"
)

var _ StringStorage = (*FileStore)(nil)

func NewFileStore(dataDir string) *FileStore {
	return &FileStore{
		DataDir: dataDir,
	}
}

// NewFileStoreAny 创建一个值类型支持泛型类型的，使用文件系统存储的 KV 存储对象
func NewFileStoreAny[V any](dataDir string, coder xcodec.Codec) *Transformer[V] {
	return &Transformer[V]{
		Codec:   coder,
		Storage: NewFileStore(dataDir),
	}
}

// FileStore 基于本地文件系统的 KV 存储实现,值类型为 string
type FileStore struct {
	// DataDir 数据存储目录，必填
	DataDir string

	// GC 触发清理多余空目录的间隔时间，可选
	// 若值 < 1秒，会使用默认值 300 秒
	GC time.Duration

	runner xpp.CooldownRunner
}

func (f *FileStore) autoCompact() {
	f.runner.Run(f.GC, f.doCompact)
}

func (f *FileStore) doCompact() {
	zos.GlobalLock()
	defer zos.GlobalUnlock()

	deleted, err := xfs.RemoveEmptyDir(f.DataDir)
	if err != nil {
		xlog.Warn(context.Background(), "anygo_xkv_FileStorage_gc", xlog.ErrorAttr("error", err))
	} else {
		xlog.Info(context.Background(), "anygo_xkv_FileStorage_gc", xlog.Int("deleted", deleted))
	}
}

func (f *FileStore) Has(ctx context.Context, key string) (bool, error) {
	fb := file.Base{
		Key: key,
		Dir: f.getDataDir(key),
	}
	info, err := fb.MetaFileStats()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func (f *FileStore) Delete(ctx context.Context, keys ...string) error {
	errs := make([]error, 0)
	for _, key := range keys {
		if err := f.deleteOne(key); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	f.autoCompact()
	return errors.Join(errs...)
}

func (f *FileStore) deleteOne(key string) error {
	fp := f.getDataDir(key)
	err := os.RemoveAll(fp)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (f *FileStore) getDataDir(key string) string {
	sg := md5.Sum([]byte("anygo" + key))
	s := hex.EncodeToString(sg[:])
	fp := filepath.Join(f.DataDir, s[:2], s[2:4], s[4:6], s[6:])
	return fp
}

func (f *FileStore) String(key string) String[string] {
	return &file.String{
		Base: &file.Base{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) List(key string) List[string] {
	return &file.List{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) Hash(key string) Hash[string] {
	return &file.Hash{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) Set(key string) Set[string] {
	return &file.Set{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) ZSet(key string) ZSet[string] {
	return &file.ZSet{
		Compact: f.autoCompact,
		Base: &file.Base{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}
