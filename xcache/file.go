//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xio"
)

var _ Cache[string, int] = (*File[string, int])(nil)

// File 文件系统缓存
type File[K comparable, V any] struct {
	// Dir 缓存文件存储目录，必填
	Dir string

	// GC 触发过期缓存清理的间隔时间，可选
	// 若值 < 1秒，会使用默认值 300 秒
	GC time.Duration

	// Codec 必填，用于数据的编解码
	Codec xcodec.Codec

	gcTime int64

	gcRunning atomic.Bool

	errChan xbus.EventBus[error]
}

func (f *File[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	select {
	case <-ctx.Done():
		return value, context.Cause(ctx)
	default:
	}
	defer f.autoGC()

	expire, data, err := f.readByKey(key, true)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return value, xerror.NotFound
		}
		return value, err
	}
	if expire {
		f.Delete(ctx, key)
		return value, xerror.NotFound
	}
	err = f.Codec.Decode(data, &value)
	return value, err
}

func (f *File[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}

	defer f.autoGC()

	fp := f.cacheFilePath(key)
	dir := filepath.Dir(fp)
	_, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(dir, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
				return err
			}
		} else {
			return err
		}
	}

	msg, err := f.Codec.Encode(value)
	if err != nil {
		return err
	}

	expireAt := time.Now().Add(ttl)

	file, err := os.CreateTemp(dir, filepath.Base(fp))
	if err != nil {
		return err
	}

	defer os.Remove(file.Name())

	// 写 cache 文件：
	writer := bufio.NewWriter(file)
	_, err = xio.WriteStrings(writer,
		// 第1行是缓存有效期，格式:etime=1590235951234907000
		"etime=",
		strconv.FormatInt(expireAt.UnixNano(), 10),
		"\n",

		// 第2行是创建时间：格式： ctime=1590235951
		"ctime=",
		strconv.FormatInt(time.Now().Unix(), 10),
		"\n",
	)

	if err == nil {
		_, err = writer.Write(msg)
	}
	if err != nil {
		return err
	}

	if err = writer.Flush(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	_ = os.Remove(fp)
	return os.Rename(file.Name(), fp)
}

func (f *File[K, V]) Delete(ctx context.Context, keys ...K) error {
	if len(keys) == 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}
	var errs []error
	for _, key := range keys {
		fp := f.cacheFilePath(key)
		err := os.Remove(fp)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (f *File[K, V]) cacheFilePath(key K) string {
	sg := md5.Sum([]byte(fmt.Sprint(key)))
	s := hex.EncodeToString(sg[:])
	fp := filepath.Join(f.Dir, s[:3], s[3:6], s[6:9], s[9:12], s[12:15], s[16:])
	return strings.Join([]string{fp, cacheFileExt}, "")
}

func (f *File[K, V]) readByKey(key K, needData bool) (expire bool, data []byte, err error) {
	fp := f.cacheFilePath(key)
	return f.readByPath(fp, needData)
}

func (f *File[K, V]) readByPath(fp string, needData bool) (expire bool, data []byte, err error) {
	info, err := os.Stat(fp)
	if err != nil {
		return false, nil, err
	}
	if info.IsDir() {
		rel, _ := filepath.Rel(f.Dir, fp)
		if rel == "" {
			rel = fp
		}
		return false, nil, fmt.Errorf("%s is directory", rel)
	}
	file, err := os.Open(fp)
	if err != nil {
		return true, nil, err
	}
	defer file.Close()

	br := bufio.NewReader(file)
	first, _, err := br.ReadLine()
	if err != nil {
		return true, nil, fmt.Errorf("read fist line : %w", err)
	}
	// 第一行为过期时间，格式为：etime=UnixNano()
	etime, ok := bytes.CutPrefix(first, []byte("etime="))
	if !ok {
		return true, nil, fmt.Errorf("not valid cache line, expect etime=\\d+, got=%q", first)
	}
	expireAt, err := strconv.ParseInt(string(etime), 10, 64)
	if err != nil {
		return true, nil, err
	}
	expire = expireAt < time.Now().UnixNano()
	if !needData {
		return expire, nil, nil
	}
	// 第二行为创建时间，格式为：ctime=unix时间戳,跳过
	_, _, err = br.ReadLine()
	if err != nil {
		return true, nil, fmt.Errorf("read second line : %w", err)
	}
	data, err = io.ReadAll(br)
	return expire, data, err
}

func (f *File[K, V]) autoGC() {
	lastGc := atomic.LoadInt64(&f.gcTime)
	newVal := time.Now().UnixNano()
	if newVal-lastGc < int64(f.getGC()) {
		return
	}

	if !atomic.CompareAndSwapInt64(&f.gcTime, lastGc, newVal) {
		return
	}
	go safely.Run(f.gc)
}

func (f *File[K, V]) getGC() time.Duration {
	if f.GC > time.Second {
		return f.GC
	}
	return 300 * time.Second
}

func (f *File[K, V]) gc() {
	if !f.gcRunning.CompareAndSwap(false, true) {
		return
	}
	defer f.gcRunning.Store(false)

	err := filepath.Walk(f.Dir, func(path string, info os.FileInfo, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if !info.IsDir() {
			if err1 := f.checkFile(path); err1 != nil && f.errChan.Subscribed() {
				f.errChan.Publish(fmt.Errorf("[xcache.File.gc] checkFile %q: %w", path, err1))
			}
		}
		return nil
	})
	if err != nil && f.errChan.Subscribed() {
		f.errChan.Publish(fmt.Errorf("[[xcache.File.gc]] Walk %q: %w", f.Dir, err))
	}
}

func (f *File[K, V]) checkFile(fp string) error {
	if strings.HasSuffix(fp, cacheFileExt) {
		return nil
	}
	expire, _, _ := f.readByPath(fp, false)
	if expire {
		return os.Remove(fp)
	}
	return nil
}
