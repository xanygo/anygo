package file

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type Set struct {
	Compact func()
	Base    *Base
}

func (f *Set) SAdd(ctx context.Context, members ...string) (int64, error) {
	if err := f.Base.SaveMeta(internal.DataTypeSet); err != nil {
		return 0, err
	}
	var added int64
	var errs []error
	for _, member := range members {
		addNew, err := f.Base.WriteKVDataFile2(f.Base.Md5(member), member)
		if err != nil {
			errs = append(errs, err)
		} else if addNew {
			added++
		}
	}
	if len(errs) == 0 {
		return added, nil
	}
	return added, errors.Join(errs...)
}

func (f *Set) SRem(ctx context.Context, members ...string) error {
	var errs []error
	for _, member := range members {
		if err := f.Base.DeleteKVDataFile(f.Base.Md5(member)); err != nil {
			errs = append(errs, err)
		}
	}
	go safely.RunVoid(f.Compact)
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// SRange 返回结果是无序的（没有按照写入顺序排序）
func (f *Set) SRange(ctx context.Context, fn func(val string) bool) error {
	err := f.Base.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		bf, err1 := os.ReadFile(filepath.Join(f.Base.Dir, d.Name()))
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

type memberWithMeta struct {
	Member string
	Mtime  int64
}

var memberSortFn = xcmp.OrderAsc(func(m memberWithMeta) int64 {
	return m.Mtime
})

// SMembers 返回所有 member，结果按照写入时间顺序正序排列
func (f *Set) SMembers(ctx context.Context) ([]string, error) {
	var list []memberWithMeta

	err := f.Base.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		bf, err1 := os.ReadFile(filepath.Join(f.Base.Dir, d.Name()))
		if err1 != nil {
			return err1
		}
		info, err2 := d.Info()
		if err2 != nil {
			return err2
		}
		list = append(list, memberWithMeta{
			Member: string(bf),
			Mtime:  info.ModTime().UnixNano(),
		})
		return nil
	})

	var result []string
	slices.SortFunc(list, memberSortFn)
	for _, m := range list {
		result = append(result, m.Member)
	}
	return result, err
}

func (f *Set) SCard(ctx context.Context) (int64, error) {
	var result int64
	err := f.Base.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		result++
		return nil
	})
	return result, err
}

func (f *Set) SIsMember(ctx context.Context, member string) (bool, error) {
	_, found, err := f.Base.CheckReadKVDataFile(f.Base.Md5(member), internal.DataTypeSet, false)
	return found, err
}

func (f *Set) SMIsMember(ctx context.Context, members []string) ([]bool, error) {
	result := make([]bool, len(members))
	for i, member := range members {
		_, found, err := f.Base.CheckReadKVDataFile(f.Base.Md5(member), internal.DataTypeSet, false)
		if err != nil {
			return nil, err
		}
		result[i] = found
	}
	return result, nil
}
