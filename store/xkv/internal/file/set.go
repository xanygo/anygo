package file

import (
	"context"
	"errors"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv/internal"
	"io/fs"
	"os"
	"path/filepath"
)

type Set struct {
	Compact func()
	Base
}

func (f Set) SAdd(ctx context.Context, members ...string) (int64, error) {
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
		return added, nil
	}
	return added, errors.Join(errs...)
}

func (f Set) SRem(ctx context.Context, members ...string) error {
	var errs []error
	for _, member := range members {
		if err := f.DeleteKVDataFile(f.Md5(member)); err != nil {
			errs = append(errs, err)
		}
	}
	go safely.RunVoid(f.Compact)
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// SRange 返回结果是无序的
func (f Set) SRange(ctx context.Context, fn func(val string) bool) error {
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

// SMembers 返回结果是无序的
func (f Set) SMembers(ctx context.Context) ([]string, error) {
	var result []string
	err := f.SRange(ctx, func(val string) bool {
		result = append(result, val)
		return true
	})
	return result, err
}

func (f Set) SCard(ctx context.Context) (int64, error) {
	var result int64
	err := f.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		result++
		return nil
	})
	return result, err
}
