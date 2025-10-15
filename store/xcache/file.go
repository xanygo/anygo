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
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/internal/fctime"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xbus"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xio"
	"github.com/xanygo/anygo/xlog"
)

var _ Cache[string, int] = (*File[string, int])(nil)
var _ MCache[string, int] = (*File[string, int])(nil)
var _ HasStats = (*File[string, int])(nil)

// File 文件系统缓存
type File[K comparable, V any] struct {
	// Dir 缓存文件存储目录，必填
	Dir string

	// GC 触发过期缓存清理的间隔时间，可选
	// 若值 < 1秒，会使用默认值 300 秒
	GC time.Duration

	// Codec 必填，用于数据的编解码
	Codec xcodec.Codec

	// Capacity Dir 目录下的最大缓存个数，>0 时有效
	// 每次 GC 时，若数量超限，会按照缓存的创建时间排序，删除创建时间更靠前的
	Capacity int

	gcTime int64

	gcRunning atomic.Bool

	errChan xbus.EventBus[error]

	readCnt   atomic.Uint64
	writeCnt  atomic.Uint64
	deleteCnt atomic.Uint64
	hitCnt    atomic.Uint64
}

func (fc *File[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	fc.readCnt.Add(1)
	defer fc.autoGC()

	select {
	case <-ctx.Done():
		return value, context.Cause(ctx)
	default:
	}
	return fc.doGet(key)
}

func (fc *File[K, V]) doGet(key K) (value V, err error) {
	expire, data, err := fc.readByKey(key, true)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return value, xerror.NotFound
		}
		return value, err
	}
	if expire {
		return value, xerror.NotFound
	}
	err = fc.Codec.Decode(data, &value)
	if err == nil {
		fc.hitCnt.Add(1)
	}
	return value, err
}

func (fc *File[K, V]) MGet(ctx context.Context, keys ...K) (map[K]V, error) {
	fc.readCnt.Add(uint64(len(keys)))
	defer fc.autoGC()

	result := make(map[K]V, len(keys))
	var errs []error
	for _, key := range keys {
		select {
		case <-ctx.Done():
			fc.hitCnt.Add(uint64(len(result)))
			return result, context.Cause(ctx)
		default:
		}

		value, err := fc.doGet(key)
		if err == nil {
			result[key] = value
			continue
		}
		if IsNotExists(err) {
			continue
		}
		errs = append(errs, err)
	}
	fc.hitCnt.Add(uint64(len(result)))
	return result, errors.Join(errs...)
}

func (fc *File[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	fc.writeCnt.Add(1)
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}

	defer fc.autoGC()
	return fc.doSet(key, value, ttl)
}

func (fc *File[K, V]) doSet(key K, value V, ttl time.Duration) error {
	fp := fc.cacheFilePath(key)
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

	msg, err := fc.Codec.Encode(value)
	if err != nil {
		return err
	}

	expireAt := time.Now().Add(ttl)

	file, err := os.CreateTemp(dir, filepath.Base(fp))
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	// 写 cache 文件：
	writer := bufio.NewWriter(file)
	_, err = xio.WriteStrings(writer,
		// 第1行是缓存有效期，格式:etime=1590235951
		"etime=",
		strconv.FormatInt(expireAt.Unix(), 10),
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

func (fc *File[K, V]) MSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	fc.writeCnt.Add(uint64(len(values)))
	defer fc.autoGC()

	var errs []error
	for k, v := range values {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		if err := fc.doSet(k, v, ttl); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (fc *File[K, V]) Delete(ctx context.Context, keys ...K) error {
	fc.deleteCnt.Add(uint64(len(keys)))
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
		fp := fc.cacheFilePath(key)
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

func (fc *File[K, V]) cacheFilePath(key K) string {
	sg := md5.Sum([]byte(fmt.Sprint(key)))
	s := hex.EncodeToString(sg[:])
	fp := filepath.Join(fc.Dir, s[:2], s[2:4], s[4:6], s[6:])
	return strings.Join([]string{fp, cacheFileExt}, "")
}

func (fc *File[K, V]) readByKey(key K, needData bool) (expire bool, data []byte, err error) {
	fp := fc.cacheFilePath(key)
	return fc.readByPath(fp, needData)
}

func (fc *File[K, V]) readByPath(fp string, needData bool) (expire bool, data []byte, err error) {
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
	// 第一行为过期时间，格式为：etime=Unix()
	etime, ok := bytes.CutPrefix(first, []byte("etime="))
	if !ok {
		return true, nil, fmt.Errorf("not valid cache line, expect etime=\\d+, got=%q", first)
	}
	expireAt, err := strconv.ParseInt(string(etime), 10, 64)
	if err != nil {
		return true, nil, err
	}
	expire = expireAt < time.Now().Unix()
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

func (fc *File[K, V]) autoGC() {
	lastGc := atomic.LoadInt64(&fc.gcTime)
	newVal := time.Now().UnixNano()
	if newVal-lastGc < int64(fc.getGC()) {
		return
	}

	if !atomic.CompareAndSwapInt64(&fc.gcTime, lastGc, newVal) {
		return
	}
	go safely.Run(fc.gc)
}

func (fc *File[K, V]) getGC() time.Duration {
	if fc.GC > time.Second {
		return fc.GC
	}
	return 300 * time.Second
}

type fileCreateTime struct {
	Path  string
	Ctime int64
}

func (fc *File[K, V]) gc() {
	if !fc.gcRunning.CompareAndSwap(false, true) {
		return
	}
	defer fc.gcRunning.Store(false)

	emptyDirs := make(map[string]bool, 10)
	var fsc []fileCreateTime
	if fc.Capacity > 0 {
		fsc = make([]fileCreateTime, 0, 100)
	}
	var fileTotal int
	start := time.Now()
	filepath.WalkDir(fc.Dir, func(path string, d fs.DirEntry, err error) error {
		if fc.Dir == path {
			return nil
		}
		fileTotal++
		parent := filepath.Dir(path)
		delete(emptyDirs, parent)

		if d.IsDir() {
			emptyDirs[path] = true
			return nil
		}

		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		expired, info, err1 := fc.checkExpire(path)
		if err1 != nil {
			fc.fireError("checkExpire", path, err1)
		}
		if !expired && info != nil && fc.Capacity > 0 {
			fsc = append(fsc, fileCreateTime{
				Path:  path,
				Ctime: fctime.Ctime(info).Unix(),
			})
		}
		return nil
	})
	walkDone := time.Now()

	logAttrs := []xlog.Attr{
		xlog.String("dir", fc.Dir),
		xlog.String("walkCost", walkDone.Sub(start).String()),
		xlog.Any("emptyDirs", emptyDirs),
		xlog.Int("fileTotal", fileTotal),
	}

	// 删除空目录，不需要往上递归，gc 方法每运行一次，会往上查找一层
	for dir := range emptyDirs {
		fc.osRemove(dir, "empty dir")
	}

	if fc.Capacity > 0 && len(fsc) > fc.Capacity {
		sort.Slice(fsc, func(i, j int) bool {
			return fsc[i].Ctime > fsc[j].Ctime
		})
		logAttrs = append(logAttrs, xlog.Int("OutOfCapacity", len(fsc)-fc.Capacity))
		for _, item := range fsc[:fc.Capacity] {
			fc.osRemove(item.Path, "by cap")
		}
	}
	logAttrs = append(logAttrs, xlog.String("TotalCost", time.Since(start).String()))
	xlog.Default().Info(context.Background(), "FileCache.GC", logAttrs...)
}

func (fc *File[K, V]) fireError(action string, fp string, err error) {
	if err == nil || !fc.errChan.Subscribed() {
		return
	}
	fc.errChan.Publish(fmt.Errorf("[[xcache.File.gc]] action=%q file=%q: %w", action, fp, err))
}

func (fc *File[K, V]) osRemove(path string, by string) {
	err := os.Remove(path)
	if err != nil {
		fc.fireError("os.Remove "+by, path, err)
	}
}

func (fc *File[K, V]) checkExpire(fp string) (expired bool, info fs.FileInfo, err error) {
	if !strings.HasSuffix(fp, cacheFileExt) {
		return false, nil, nil
	}
	file, err := os.Open(fp)
	if err != nil {
		return true, nil, err
	}
	defer file.Close()

	info, err = file.Stat()
	if err != nil {
		fc.osRemove(fp, "Stat failed")
		return true, nil, err
	}

	br := bufio.NewReader(file)
	first, _, err := br.ReadLine()
	if err != nil {
		fc.osRemove(fp, "ReadLine failed")
		return true, info, fmt.Errorf("read fist line : %w", err)
	}
	// 第一行为过期时间，格式为：etime=Unix()
	etime, ok := bytes.CutPrefix(first, []byte("etime="))
	if !ok {
		fc.osRemove(fp, "invalid content")
		return true, nil, fmt.Errorf("not valid cache line, expect etime=\\d+, got=%q", first)
	}
	expireAt, err := strconv.ParseInt(string(etime), 10, 64)
	if err != nil {
		fc.osRemove(fp, "invalid expire time")
		return true, nil, err
	}
	expired = expireAt < time.Now().Unix()
	if expired {
		err = os.Remove(fp)
	}
	return expired, info, err
}

func (fc *File[K, V]) Stats() Stats {
	return Stats{
		Read:   fc.readCnt.Load(),
		Write:  fc.writeCnt.Load(),
		Delete: fc.deleteCnt.Load(),
		Hit:    fc.hitCnt.Load(),
	}
}
