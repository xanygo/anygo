package file

import (
	"context"
	"errors"
	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv/internal"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type List struct {
	Compact func()
	Base
}

// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
func (f List) LPush(ctx context.Context, values ...string) (int64, error) {
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

func (f List) RPush(ctx context.Context, values ...string) (int64, error) {
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
func (f List) LPop(ctx context.Context) (string, bool, error) {
	return f.pop(ctx, true)
}

func (f List) pop(ctx context.Context, left bool) (string, bool, error) {
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
	value, ok, err := f.ReadFile(fileName, true)
	go safely.RunVoid(f.Compact)
	return value, ok, err
}

func (f List) compare(a string, b string) bool {
	return a > b
}

func (f List) RPop(ctx context.Context) (string, bool, error) {
	return f.pop(ctx, false)
}

func (f List) LRem(ctx context.Context, count int64, element string) (deleted int64, err error) {
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

type fileNameInfo struct {
	Name     string
	Flag     int   // 0 或 1，0-LPush 1-RPush
	Timespan int64 // 时间戳
}

// LRange 查询数据的排序算法
var fileNameSortAsc = xcmp.Chain(
	xcmp.TrueFront[fileNameInfo](func(info fileNameInfo) bool {
		return info.Flag == 0
	}),
	xcmp.OrderAsc[fileNameInfo, int64](func(info fileNameInfo) int64 {
		return info.Timespan
	}),
)

// RRange 查询数据的排序算法
var fileNameSortDesc = xcmp.Chain(
	xcmp.TrueFront[fileNameInfo](func(info fileNameInfo) bool {
		return info.Flag == 1
	}),
	xcmp.OrderDesc[fileNameInfo, int64](func(info fileNameInfo) int64 {
		return info.Timespan
	}),
)

func (f List) lrRange(ctx context.Context, left bool, fn func(path, val string) bool) error {
	var fileInfos []fileNameInfo
	err := f.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		flag, timespan := f.parserKVDFileName(d.Name())
		fileInfos = append(fileInfos, fileNameInfo{
			Name:     d.Name(),
			Flag:     flag,
			Timespan: timespan,
		})
		return nil
	})

	if err != nil {
		return err
	}

	if left {
		slices.SortFunc(fileInfos, fileNameSortAsc)
	} else {
		slices.SortFunc(fileInfos, fileNameSortDesc)
	}

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

func (f List) parserKVDFileName(name string) (int, int64) {
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
	if flag == 0 { // 0-LPUSH,时间越大排名越靠前
		timespan *= -1
	}
	return flag, timespan
}

func (f List) LRange(ctx context.Context, fn func(val string) bool) error {
	return f.lrRange(ctx, true, func(path, val string) bool {
		return fn(val)
	})
}

func (f List) RRange(ctx context.Context, fn func(val string) bool) error {
	return f.lrRange(ctx, false, func(path, val string) bool {
		return fn(val)
	})
}

// Range 无序的
func (f List) Range(ctx context.Context, fn func(val string) bool) error {
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

func (f List) LLen(ctx context.Context) (int64, error) {
	var num int64
	err := f.Range(ctx, func(val string) bool {
		num++
		return true
	})
	return num, err
}
