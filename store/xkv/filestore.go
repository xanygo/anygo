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

	"github.com/xanygo/anygo/store/xkv/internal"
)

var ErrInvalidType = internal.ErrInvalidType

var _ Storage = (*FileStorage)(nil)

// FileStorage 基于本地文件系统的 KV 存储实现
type FileStorage struct {
	// DataDir 数据存储目录，必填
	DataDir string
}

func (f *FileStorage) Delete(ctx context.Context, key string) error {
	fp := f.getDataDir(key)
	err := os.RemoveAll(fp)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (f *FileStorage) getDataDir(key string) string {
	sg := md5.Sum([]byte("anygo" + key))
	s := hex.EncodeToString(sg[:])
	fp := filepath.Join(f.DataDir, s[:3], s[3:6], s[6:9], s[9:12], s[12:15], s[16:])
	return fp
}

func (f *FileStorage) String(key string) String {
	return &fileString{
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStorage) List(key string) List {
	return &fileList{
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStorage) Hash(key string) Hash {
	return &fileHash{
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStorage) Set(key string) Set {
	return &fileSet{
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

func (f *FileStorage) ZSet(key string) ZSet {
	return &fileZSet{
		FileBase: internal.FileBase{
			Key: key,
			Dir: f.getDataDir(key),
		},
	}
}

var _ String = (*fileString)(nil)

type fileString struct {
	internal.FileBase
}

func (f *fileString) Set(ctx context.Context, value string) error {
	if err := f.SaveMeta(internal.DataTypeString); err != nil {
		return err
	}
	return f.WriteKVDataFile("value", value)
}

func (f *fileString) Get(ctx context.Context) (string, error) {
	content, _, err := f.CheckReadKVDataFile("value", internal.DataTypeString, false)
	return content, err
}

func (f *fileString) Incr(ctx context.Context) (int64, error) {
	value, err := f.Get(ctx)
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
	value, err := f.Get(ctx)
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

var _ List = (*fileList)(nil)

type fileList struct {
	internal.FileBase
}

// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
func (f fileList) LPush(ctx context.Context, val string) error {
	if err := f.SaveMeta(internal.DataTypeList); err != nil {
		return err
	}

	name := strconv.FormatInt(time.Now().UnixNano(), 10)
	return f.WriteKVDataFile("0_"+name, val)
}

func (f fileList) RPush(ctx context.Context, val string) error {
	if err := f.SaveMeta(internal.DataTypeList); err != nil {
		return err
	}

	name := strconv.FormatInt(time.Now().UnixNano(), 10)
	return f.WriteKVDataFile("1_"+name, val)
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
	return f.ReadFile(fileName, true)
}

func (f fileList) compare(a string, b string) bool {
	return a > b
}

func (f fileList) RPop(ctx context.Context) (string, bool, error) {
	return f.pop(ctx, false)
}

func (f fileList) lrRange(ctx context.Context, left bool, fn func(val string) bool) error {
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
		bf, err := os.ReadFile(filepath.Join(f.Dir, fileInfo.Name))
		if err != nil {
			return err
		}
		if !fn(string(bf)) {
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
	return f.lrRange(ctx, true, fn)
}

func (f fileList) RRange(ctx context.Context, fn func(val string) bool) error {
	return f.lrRange(ctx, false, fn)
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

var _ Hash = (*fileHash)(nil)

type fileHash struct {
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

func (f fileHash) HDel(ctx context.Context, field string) error {
	return f.DeleteKVDataFile(f.Md5(field))
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

var _ Set = (*fileSet)(nil)

type fileSet struct {
	internal.FileBase
}

func (f fileSet) SAdd(ctx context.Context, val string) error {
	if err := f.SaveMeta(internal.DataTypeSet); err != nil {
		return err
	}
	return f.WriteKVDataFile(f.Md5(val), val)
}

func (f fileSet) SRem(ctx context.Context, val string) error {
	return f.DeleteKVDataFile(f.Md5(val))
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

var _ ZSet = (*fileZSet)(nil)

type fileZSet struct {
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

func (f fileZSet) ZRem(ctx context.Context, member string) error {
	return f.DeleteKVDataFile(f.Md5(member))
}

type fileZSetMember struct {
	Member string  `json:"m"`
	Score  float64 `json:"s"`
}
