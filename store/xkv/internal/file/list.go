package file

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/safely"
)

type List struct {
	Compact func()
	Base    *Base
}

// LPush 在列表左侧插入元素（类似 Redis 的 LPUSH 命令）
func (l *List) LPush(ctx context.Context, values ...string) (int64, error) {
	if len(values) > 0 {
		err := l.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
			id := time.Now().UnixNano()
			for _, value := range values {
				name := strconv.FormatInt(id, 10)
				_, err1 := l.Base.writeMemberFile2("0_"+name, value)
				if err1 != nil {
					return err1
				}
				id++
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	return l.LLen(ctx)
}

func (l *List) RPush(ctx context.Context, values ...string) (int64, error) {
	if len(values) > 0 {
		err := l.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
			id := time.Now().UnixNano()
			for _, value := range values {
				name := strconv.FormatInt(id, 10)
				_, err1 := l.Base.writeMemberFile2("1_"+name, value)
				if err1 != nil {
					return err1
				}
				id++
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	return l.LLen(ctx)
}

// LPop 移除并返回列表最左侧的元素（类似 Redis 的 LPOP 命令）
func (l *List) LPop(ctx context.Context) (string, bool, error) {
	return l.pop(ctx, true)
}

func (l *List) pop(ctx context.Context, left bool) (value string, found bool, err error) {
	err = l.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		defer l.Base.deleteKeyWhenNoMember(ctx)

		var fileName string
		err1 := l.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			if fileName == "" {
				fileName = path
			} else if l.compare(path, fileName) == left {
				fileName = path
			}
			return nil
		})

		if err1 != nil || fileName == "" {
			return err1
		}
		v, ok, err2 := l.Base.readFile(fileName, true)
		if err2 != nil || !ok {
			return err2
		}
		value = v
		found = true

		return nil
	})
	go safely.RunVoid(l.Compact)
	return value, found, err
}

func (l *List) compare(a string, b string) bool {
	return a > b
}

func (l *List) RPop(ctx context.Context) (string, bool, error) {
	return l.pop(ctx, false)
}

func (l *List) LRem(ctx context.Context, count int64, element string) (deleted int64, err error) {
	defer func() {
		_ = l.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
			l.Base.deleteKeyWhenNoMember(ctx)
			return nil
		})
	}()
	callBack := func(path, val string) (bool, error) {
		if val != element {
			return true, nil
		}
		err1 := l.Base.osRemove(path)
		if err1 != nil {
			return false, err1
		}
		deleted++
		return deleted < count, nil
	}
	if count >= 0 {
		err = l.lrRange(ctx, true, true, callBack)
	} else {
		count = count * -1
		err = l.lrRange(ctx, true, false, callBack)
	}
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

type listFileNameInfo struct {
	Path     string
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

func (l *List) lrRange(ctx context.Context, write bool, left bool, fn func(path, val string) (bool, error)) error {
	callBack := func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		var fileInfos []listFileNameInfo
		err1 := l.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			flag, timespan := l.parserKVDFileName(d.Name())
			fileInfos = append(fileInfos, listFileNameInfo{
				Path:     path,
				Flag:     flag,
				Timespan: timespan,
			})
			return nil
		})
		if err1 != nil {
			return err1
		}
		if left {
			slices.SortFunc(fileInfos, listFileNameSortAsc)
		} else {
			slices.SortFunc(fileInfos, listFileNameSortDesc)
		}

		for _, fileInfo := range fileInfos {
			bf, err2 := os.ReadFile(fileInfo.Path)
			if err2 != nil {
				return err2
			}
			ok, err3 := fn(fileInfo.Path, string(bf))
			if !ok || err3 != nil {
				return err3
			}
		}
		return nil
	}
	if write {
		return l.Base.lock(ctx, callBack)
	}
	return l.Base.lockRead(ctx, callBack)
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
	return l.lrRange(ctx, false, true, func(path, val string) (bool, error) {
		return fn(val), nil
	})
}

func (l *List) RRange(ctx context.Context, fn func(val string) bool) error {
	return l.lrRange(ctx, false, false, func(path, val string) (bool, error) {
		return fn(val), nil
	})
}

// Range 无序的
func (l *List) Range(ctx context.Context, fn func(val string) bool) error {
	return l.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return l.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			bf, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !fn(string(bf)) {
				return fs.SkipAll
			}
			return nil
		})
	})
}

func (l *List) LLen(ctx context.Context) (int64, error) {
	var num int64
	err := l.Range(ctx, func(val string) bool {
		num++
		return true
	})
	return num, err
}
