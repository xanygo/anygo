//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package file

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/xio/xfs"
)

const memberFileExt = ".kvd"

type Meta struct {
	Key     string            `json:"k"`
	Type    internal.DataType `json:"t"`
	Created int64             `json:"c"`
	Updated int64             `json:"u"`
}

type Base struct {
	Key        string
	Dir        string
	Type       internal.DataType
	GroupMutex *xsync.GroupMutex[any]
}

func (fb *Base) metaOrNew(meta *Meta) *Meta {
	if meta != nil {
		return meta
	}
	now := time.Now().Unix()
	return &Meta{
		Key:     fb.Key,
		Type:    fb.Type,
		Created: now,
		Updated: now,
	}
}

func (fb *Base) getMetaFilePath() string {
	return filepath.Join(fb.Dir, "meta")
}

func (fb *Base) deleteKey() error {
	err := os.RemoveAll(fb.Dir)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// deleteKeyWhenNoMember 当没有 member 或者 fields 的时候，删除key
func (fb *Base) deleteKeyWhenNoMember(ctx context.Context) {
	ok, err := fb.hasMembers(ctx)
	if ok || err != nil {
		return
	}
	_ = fb.deleteKey()
}

func (fb *Base) hasMembers(ctx context.Context) (ok bool, err error) {
	err = fb.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
		ok = true
		return fs.SkipAll
	})
	return ok, err
}

func (fb *Base) readMeta() (*Meta, error) {
	bf, err := os.ReadFile(fb.getMetaFilePath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	meta := &Meta{}
	err = json.Unmarshal(bf, meta)
	return meta, err
}

func (fb *Base) doWithMeta(ctx context.Context, fn func(ctx context.Context, meta *Meta) error) error {
	meta, err := fb.readMeta()
	if err != nil {
		return err
	}
	if meta != nil && !meta.Type.Equal(fb.Type) {
		return fmt.Errorf("%w, cannot read/write %s on %s", internal.ErrInvalidType, fb.Type, meta.Type)
	}
	return fn(ctx, meta)
}

// lock 实现写锁，直接将读到的 meta 传给回调方法
func (fb *Base) lock(ctx context.Context, fn func(ctx context.Context, meta *Meta) error) error {
	mux := fb.GroupMutex.Locker(fb.Key)
	mux.Lock()
	defer mux.Unlock()
	return fb.doWithMeta(ctx, fn)
}

// lockWrite 使用写锁，并且若 meta 不存在，会线写 meta，然后再执行回调
func (fb *Base) lockWrite(ctx context.Context, fn func(ctx context.Context, meta *Meta) error) error {
	mux := fb.GroupMutex.Locker(fb.Key)
	mux.Lock()
	defer mux.Unlock()

	return fb.doWithMeta(ctx, func(ctx context.Context, meta *Meta) error {
		meta = fb.metaOrNew(meta)
		if err := fb.saveMeta(meta); err != nil {
			return err
		}
		return fn(ctx, meta)
	})
}

// lockRead 使用读锁
func (fb *Base) lockRead(ctx context.Context, fn func(ctx context.Context, meta *Meta) error) error {
	mux := fb.GroupMutex.Locker(fb.Key)
	mux.RLock()
	defer mux.RUnlock()

	return fb.doWithMeta(ctx, fn)
}

func (fb *Base) saveMeta(meta *Meta) error {
	if err := xfs.KeepDirExists(fb.Dir); err != nil {
		return err
	}
	bf, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	fp := fb.getMetaFilePath()
	return os.WriteFile(fp, bf, 0644)
}

// writeMemberFile2 写入，返回：(新写入/更新，错误)
func (fb *Base) writeMemberFile2(baseName string, data string) (added bool, err error) {
	baseName = baseName + memberFileExt
	fp := filepath.Join(fb.Dir, baseName)
	dir := filepath.Dir(fp)
	if err = xfs.KeepDirExists(dir); err != nil {
		return false, err
	}
	bf := unsafe.Slice(unsafe.StringData(data), len(data))
	old, err := os.ReadFile(fp)
	if err == nil {
		if bytes.Equal(old, bf) {
			// 文件内容没有变化
			return false, nil
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	// 是否新增
	return true, os.WriteFile(fp, bf, 0644)
}

func (fb *Base) writeFile(baseName string, data string) error {
	fp := filepath.Join(fb.Dir, baseName)
	dir := filepath.Dir(fp)
	if err := xfs.KeepDirExists(dir); err != nil {
		return err
	}
	bf := unsafe.Slice(unsafe.StringData(data), len(data))
	return os.WriteFile(fp, bf, 0644)
}

// writeMemberFile 写 kv 数据文件
func (fb *Base) writeMemberFile(baseName string, data string) error {
	return fb.writeFile(baseName+memberFileExt, data)
}

func (fb *Base) readMemberFile(baseName string) (string, bool, error) {
	return fb.readFile(baseName+memberFileExt, false)
}

func (fb *Base) readMemberFileByPath(fp string) ([]byte, error) {
	bf, err := os.ReadFile(fp)
	if err == nil {
		return bf, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, err
}

func (fb *Base) readFile(baseName string, delete bool) (string, bool, error) {
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

func (fb *Base) deleteMemberFile(baseName string) error {
	return fb.deleteFile(baseName + memberFileExt)
}

func (fb *Base) deleteFile(baseName string) error {
	fp := filepath.Join(fb.Dir, baseName)
	return fb.osRemove(fp)
}

func (fb *Base) osRemove(fp string) error {
	err := os.Remove(fp)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (fb *Base) md5(field string) string {
	sm := md5.Sum([]byte("anygo" + field))
	return hex.EncodeToString(sm[:])
}

func (fb *Base) rangeMemberFiles(ctx context.Context, fn func(path string, d fs.DirEntry) error) error {
	err := fs.WalkDir(os.DirFS(fb.Dir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
			// 继续
		}

		if !strings.HasSuffix(d.Name(), memberFileExt) {
			return nil
		}
		err = fn(filepath.Join(fb.Dir, path), d)
		if err != nil && errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (fb *Base) Has(ctx context.Context) (ok bool, err error) {
	err = fb.lock(ctx, func(ctx context.Context, meta *Meta) error {
		ok = meta != nil
		return nil
	})
	return ok, err
}
