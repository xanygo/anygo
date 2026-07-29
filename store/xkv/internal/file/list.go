package file

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type List struct {
	Compact func()
	Base    *Base
}

// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
func (l *List) LPush(ctx context.Context, values ...string) (int64, error) {
	if err := l.Base.SaveMeta(internal.DataTypeList); err != nil {
		return 0, err
	}
	var errs []error
	id := time.Now().UnixNano()
	for _, value := range values {
		name := strconv.FormatInt(id, 10)
		_, err := l.Base.WriteKVDataFile2("0_"+name, value)
		id++
		if err != nil {
			errs = append(errs, err)
		}
	}
	num, err := l.LLen(ctx)
	if err != nil {
		errs = append(errs, err)
	}
	return num, errors.Join(errs...)
}

func (l *List) RPush(ctx context.Context, values ...string) (int64, error) {
	if err := l.Base.SaveMeta(internal.DataTypeList); err != nil {
		return 0, err
	}

	var errs []error
	id := time.Now().UnixNano()
	for _, value := range values {
		name := strconv.FormatInt(id, 10)
		_, err := l.Base.WriteKVDataFile2("1_"+name, value)
		id++
		if err != nil {
			errs = append(errs, err)
		}
	}
	num, err := l.LLen(ctx)
	if err != nil {
		errs = append(errs, err)
	}
	return num, errors.Join(errs...)
}

// LPop 移除并返回列表最左侧的元素（类似 Redis 的 LPOP 命令）
func (l *List) LPop(ctx context.Context) (string, bool, error) {
	return l.pop(ctx, true)
}

func (l *List) pop(ctx context.Context, left bool) (string, bool, error) {
	var fileName string
	err := l.Base.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		if fileName == "" {
			fileName = path
		} else if l.compare(path, fileName) == left {
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
	value, ok, err := l.Base.ReadFile(fileName, true)
	go safely.RunVoid(l.Compact)
	return value, ok, err
}

func (l *List) compare(a string, b string) bool {
	return a > b
}

func (l *List) RPop(ctx context.Context) (string, bool, error) {
	return l.pop(ctx, false)
}

func (l *List) LRem(ctx context.Context, count int64, element string) (deleted int64, err error) {
	var errs []error
	callBack := func(path, val string) bool {
		if val != element {
			return true
		}
		if err1 := l.Base.OsRemove(path); err1 == nil {
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
		err = l.lrRange(ctx, true, callBack)
	} else {
		count = count * -1
		err = l.lrRange(ctx, false, callBack)
	}
	if err != nil {
		errs = append(errs, err)
	}
	return deleted, errors.Join(errs...)
}

type listFileNameInfo struct {
	Name     string
	Flag     int   // 0 或 1，0-LPush 1-RPush
	Timespan int64 // 时间戳
}

// LRange 查询数据的排序算法
var listFileNameSortAsc = xcmp.Chain(
	xcmp.TrueFront[listFileNameInfo](func(info listFileNameInfo) bool {
		return info.Flag == 0
	}),
	xcmp.OrderAsc[listFileNameInfo, int64](func(info listFileNameInfo) int64 {
		return info.Timespan
	}),
)

// RRange 查询数据的排序算法
var listFileNameSortDesc = xcmp.Chain(
	xcmp.TrueFront[listFileNameInfo](func(info listFileNameInfo) bool {
		return info.Flag == 1
	}),
	xcmp.OrderDesc[listFileNameInfo, int64](func(info listFileNameInfo) int64 {
		return info.Timespan
	}),
)

func (l *List) lrRange(ctx context.Context, left bool, fn func(path, val string) bool) error {
	var fileInfos []listFileNameInfo
	err := l.Base.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		flag, timespan := l.parserKVDFileName(d.Name())
		fileInfos = append(fileInfos, listFileNameInfo{
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
		slices.SortFunc(fileInfos, listFileNameSortAsc)
	} else {
		slices.SortFunc(fileInfos, listFileNameSortDesc)
	}

	for _, fileInfo := range fileInfos {
		fp := filepath.Join(l.Base.Dir, fileInfo.Name)
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

func (l *List) parserKVDFileName(name string) (int, int64) {
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

func (l *List) LRange(ctx context.Context, fn func(val string) bool) error {
	return l.lrRange(ctx, true, func(path, val string) bool {
		return fn(val)
	})
}

func (l *List) RRange(ctx context.Context, fn func(val string) bool) error {
	return l.lrRange(ctx, false, func(path, val string) bool {
		return fn(val)
	})
}

// Range 无序的
func (l *List) Range(ctx context.Context, fn func(val string) bool) error {
	err := l.Base.RangeKVFiles(ctx, internal.DataTypeList, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(l.Base.Dir, d.Name()))
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

func (l *List) LLen(ctx context.Context) (int64, error) {
	var num int64
	err := l.Range(ctx, func(val string) bool {
		num++
		return true
	})
	return num, err
}
