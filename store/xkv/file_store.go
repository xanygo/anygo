//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xanygo/anygo/internal/zos"
	"github.com/xanygo/anygo/store/xkv/internal"
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

func (f *FileStore) autoGC() {
	f.runner.Run(f.GC, f.doGC)
}

func (f *FileStore) doGC() {
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
	fb := internal.FileBase{
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
	f.autoGC()
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
	return &fileString{
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) List(key string) List[string] {
	return &fileList{
		fss: f,
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) Hash(key string) Hash[string] {
	return &fileHash{
		fss: f,
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) Set(key string) Set[string] {
	return &fileSet{
		fss: f,
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStore) ZSet(key string) ZSet[string] {
	return &fileZSet{
		fss: f,
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

var _ String[string] = (*fileString)(nil)

type fileString struct {
	internal.FileBase
}

func (f *fileString) Set(ctx context.Context, value string) error {
	if err := f.SaveMeta(internal.DataTypeString); err != nil {
		return err
	}
	return f.WriteKVDataFile("value", value)
}

func (f *fileString) Get(ctx context.Context) (string, bool, error) {
	return f.CheckReadKVDataFile("value", internal.DataTypeString, false)
}

func (f *fileString) Incr(ctx context.Context) (int64, error) {
	value, _, err := f.Get(ctx)
	if err != nil {
		return 0, err
	}
	num, _ := strconv.ParseInt(value, 10, 64)
	num++
	err = f.Set(ctx, strconv.FormatInt(num, 10))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (f *fileString) Decr(ctx context.Context) (int64, error) {
	value, _, err := f.Get(ctx)
	if err != nil {
		return 0, err
	}
	num, _ := strconv.ParseInt(value, 10, 64)
	num--
	err = f.Set(ctx, strconv.FormatInt(num, 10))
	if err != nil {
		return 0, err
	}
	return num, nil
}

var _ List[string] = (*fileList)(nil)

type fileList struct {
	fss *FileStore
	internal.FileBase
}

// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
func (f fileList) LPush(ctx context.Context, values ...string) (int64, error) {
	if err := f.SaveMeta(internal.DataTypeList); err != nil {
		return 0, err
	}
	var errs []error
	id := time.Now().UnixNano()
	for _, value := range values {
		name := strconv.FormatInt(id, 10)
		_, err := f.WriteKVDataFile2("0_"+name, value)
		id++
		if err != nil {
			errs = append(errs, err)
		}
	}
	num, err := f.LLen(ctx)
	if err != nil {
		errs = append(errs, err)
	}
	return num, errors.Join(errs...)
}

func (f fileList) RPush(ctx context.Context, values ...string) (int64, error) {
	if err := f.SaveMeta(internal.DataTypeList); err != nil {
		return 0, err
	}

	var errs []error
	id := time.Now().UnixNano()
	for _, value := range values {
		name := strconv.FormatInt(id, 10)
		_, err := f.WriteKVDataFile2("1_"+name, value)
		id++
		if err != nil {
			errs = append(errs, err)
		}
	}
	num, err := f.LLen(ctx)
	if err != nil {
		errs = append(errs, err)
	}
	return num, errors.Join(errs...)
}

// LPop 移除并返回列表最左侧的元素（类似 Redis 的 LPOP 命令）
func (f fileList) LPop(ctx context.Context) (string, bool, error) {
	return f.pop(ctx, true)
}

func (f fileList) pop(ctx context.Context, left bool) (string, bool, error) {
	var fileName string
	err := f.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		if fileName == "" {
			fileName = path
		} else if f.compare(path, fileName) == left {
			fileName = path
		}
		return nil
	})
	if err != nil {
		return "", false, err
	}
	if fileName == "" {
		return "", false, nil
	}
	f.fss.autoGC()
	return f.ReadFile(fileName, true)
}

func (f fileList) compare(a string, b string) bool {
	return a > b
}

func (f fileList) RPop(ctx context.Context) (string, bool, error) {
	return f.pop(ctx, false)
}

func (f fileList) LRem(ctx context.Context, count int64, element string) (deleted int64, err error) {
	var errs []error
	callBack := func(path, val string) bool {
		if val != element {
			return true
		}
		if err1 := f.OsRemove(path); err1 == nil {
			errs = append(errs, err1)
		} else {
			deleted++
			if count > 0 && deleted >= count {
				return false
			}
		}
		return true
	}
	if count >= 0 {
		err = f.lrRange(ctx, true, callBack)
	} else {
		count = count * -1
		err = f.lrRange(ctx, false, callBack)
	}
	if err != nil {
		errs = append(errs, err)
	}
	return deleted, errors.Join(errs...)
}

func (f fileList) lrRange(ctx context.Context, left bool, fn func(path, val string) bool) error {
	type fileNameInfo struct {
		Name     string
		Flag     int
		Timespan int64
	}
	var fileInfos []fileNameInfo
	err := f.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		flag, timespan := f.parserKVDFileName(d.Name())
		if timespan > 0 {
			fileInfos = append(fileInfos, fileNameInfo{
				Name:     d.Name(),
				Flag:     flag,
				Timespan: timespan,
			})
		}
		return nil
	})

	if err != nil {
		return err
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		a := fileInfos[i]
		b := fileInfos[j]
		result := a.Flag <= b.Flag && a.Timespan < b.Timespan
		if left {
			return result
		}
		return !result
	})
	for _, fileInfo := range fileInfos {
		fp := filepath.Join(f.Dir, fileInfo.Name)
		bf, err := os.ReadFile(fp)
		if err != nil {
			return err
		}
		if !fn(fp, string(bf)) {
			return nil
		}
	}
	return nil
}

func (f fileList) parserKVDFileName(name string) (int, int64) {
	name, found := strings.CutSuffix(name, filepath.Ext(name))
	if !found {
		return 0, 0
	}
	before, after, found := strings.Cut(name, "_")
	if !found {
		return 0, 0
	}
	flag, err := strconv.Atoi(before)
	if err != nil {
		return 0, 0
	}
	timespan, err := strconv.ParseInt(after, 10, 64)
	if err != nil {
		return 0, 0
	}
	return flag, timespan
}

func (f fileList) LRange(ctx context.Context, fn func(val string) bool) error {
	return f.lrRange(ctx, true, func(path, val string) bool {
		return fn(val)
	})
}

func (f fileList) RRange(ctx context.Context, fn func(val string) bool) error {
	return f.lrRange(ctx, false, func(path, val string) bool {
		return fn(val)
	})
}

// Range 无序的
func (f fileList) Range(ctx context.Context, fn func(val string) bool) error {
	err := f.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(f.Dir, d.Name()))
		if err != nil {
			return err
		}
		if !fn(string(bf)) {
			return fs.SkipAll
		}
		return nil
	})
	return err
}

func (f fileList) LLen(ctx context.Context) (int64, error) {
	var num int64
	err := f.Range(ctx, func(val string) bool {
		num++
		return true
	})
	return num, err
}

var _ Hash[string] = (*fileHash)(nil)

type fileHash struct {
	fss *FileStore
	internal.FileBase
}

type fileHashKV struct {
	Field string `json:"f"`
	Value string `json:"v"`
}

func (f fileHashKV) String() string {
	bf, _ := json.Marshal(f)
	return string(bf)
}

func (f fileHash) HSet(ctx context.Context, field, value string) error {
	if err := f.SaveMeta(internal.DataTypeHash); err != nil {
		return err
	}
	kv := fileHashKV{
		Field: field,
		Value: value,
	}
	return f.WriteKVDataFile(f.Md5(field), kv.String())
}

func (f fileHash) HMSet(ctx context.Context, values map[string]string) error {
	if err := f.SaveMeta(internal.DataTypeHash); err != nil {
		return err
	}
	var errs []error
	for k, v := range values {
		kv := fileHashKV{
			Field: k,
			Value: v,
		}
		if err := f.WriteKVDataFile(f.Md5(k), kv.String()); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (f fileHash) HGet(ctx context.Context, field string) (string, bool, error) {
	str, found, err := f.CheckReadKVDataFile(f.Md5(field), internal.DataTypeHash, false)
	if err != nil || !found {
		return "", false, err
	}
	kv := &fileHashKV{}
	err = json.Unmarshal([]byte(str), kv)
	if err != nil {
		return "", false, err
	}
	return kv.Value, true, nil
}

func (f fileHash) HDel(ctx context.Context, fields ...string) error {
	var errs []error
	for _, field := range fields {
		if err := f.DeleteKVDataFile(f.Md5(field)); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	f.fss.autoGC()
	return errors.Join(errs...)
}

func (f fileHash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	err := f.RangeKVFiles(ctx, internal.DataTypeHash, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(f.Dir, d.Name()))
		if err != nil {
			return err
		}
		kv := &fileHashKV{}
		err = json.Unmarshal(bf, kv)
		if err != nil {
			return err
		}
		if !fn(kv.Field, kv.Value) {
			return fs.SkipAll
		}
		return nil
	})
	return err
}

func (f fileHash) HGetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	err := f.HRange(ctx, func(field string, value string) bool {
		result[field] = value
		return true
	})
	return result, err
}

var _ Set[string] = (*fileSet)(nil)

type fileSet struct {
	fss *FileStore
	internal.FileBase
}

func (f fileSet) SAdd(ctx context.Context, members ...string) (int64, error) {
	if err := f.SaveMeta(internal.DataTypeSet); err != nil {
		return 0, err
	}
	var added int64
	var errs []error
	for _, member := range members {
		addNew, err := f.WriteKVDataFile2(f.Md5(member), member)
		if err != nil {
			errs = append(errs, err)
		} else if addNew {
			added++
		}
	}
	if len(errs) == 0 {
		return int64(len(members)), nil
	}
	return added, errors.Join(errs...)
}

func (f fileSet) SRem(ctx context.Context, members ...string) error {
	var errs []error
	for _, member := range members {
		if err := f.DeleteKVDataFile(f.Md5(member)); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	f.fss.autoGC()
	return errors.Join(errs...)
}

func (f fileSet) SRange(ctx context.Context, fn func(val string) bool) error {
	err := f.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		bf, err1 := os.ReadFile(filepath.Join(f.Dir, d.Name()))
		if err1 != nil {
			return err1
		}
		if !fn(string(bf)) {
			return fs.SkipAll
		}
		return nil
	})
	return err
}

func (f fileSet) SMembers(ctx context.Context) ([]string, error) {
	var result []string
	err := f.SRange(ctx, func(val string) bool {
		result = append(result, val)
		return true
	})
	return result, err
}

func (f fileSet) SCard(ctx context.Context) (int64, error) {
	var result int64
	err := f.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		result++
		return nil
	})
	return result, err
}

var _ ZSet[string] = (*fileZSet)(nil)

type fileZSet struct {
	fss *FileStore
	internal.FileBase
}

func (f fileZSet) ZAdd(ctx context.Context, score float64, member string) error {
	if err := f.SaveMeta(internal.DataTypeZSet); err != nil {
		return err
	}
	m := fileZSetMember{
		Member: member,
		Score:  score,
	}
	bf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return f.WriteKVDataFile(f.Md5(member), string(bf))
}

func (f fileZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	str, found, err := f.CheckReadKVDataFile(f.Md5(member), internal.DataTypeZSet, false)
	if err != nil || !found {
		return 0, false, err
	}
	m := &fileZSetMember{}
	err = json.Unmarshal([]byte(str), m)
	return m.Score, err == nil, err
}

func (f fileZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	var list []*fileZSetMember
	err := f.RangeKVFiles(ctx, internal.DataTypeZSet, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(f.Dir, d.Name()))
		if err != nil {
			return err
		}
		m := &fileZSetMember{}
		err = json.Unmarshal(bf, m)
		if err == nil {
			list = append(list, m)
		}
		return err
	})
	if err != nil {
		return err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score < list[j].Score
	})
	for _, m := range list {
		if !fn(m.Member, m.Score) {
			return nil
		}
	}
	return err
}

func (f fileZSet) ZRem(ctx context.Context, members ...string) error {
	var errs []error
	for _, member := range members {
		if err := f.DeleteKVDataFile(f.Md5(member)); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	f.fss.autoGC()
	return errors.Join(errs...)
}

type fileZSetMember struct {
	Member string  `json:"m"`
	Score  float64 `json:"s"`
}
