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

func (s *Set) SAdd(ctx context.Context, members ...string) (int64, error) {
	if err := s.Base.SaveMeta(internal.DataTypeSet); err != nil {
		return 0, err
	}
	var added int64
	for _, member := range members {
		addNew, err := s.Base.WriteKVDataFile2(s.Base.Md5(member), member)
		if err != nil {
			return 0, err
		}
		if addNew {
			added++
		}
	}
	return added, nil
}

func (s *Set) SRem(ctx context.Context, members ...string) error {
	var errs []error
	for _, member := range members {
		if err := s.Base.DeleteKVDataFile(s.Base.Md5(member)); err != nil {
			errs = append(errs, err)
		}
	}
	go safely.RunVoid(s.Compact)
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// SRange 返回结果是无序的（没有按照写入顺序排序）
func (s *Set) SRange(ctx context.Context, fn func(val string) bool) error {
	err := s.Base.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		bf, err1 := os.ReadFile(filepath.Join(s.Base.Dir, d.Name()))
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
func (s *Set) SMembers(ctx context.Context) ([]string, error) {
	var list []memberWithMeta

	err := s.Base.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		bf, err1 := os.ReadFile(filepath.Join(s.Base.Dir, d.Name()))
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

func (s *Set) SCard(ctx context.Context) (int64, error) {
	var result int64
	err := s.Base.RangeKVFiles(ctx, internal.DataTypeSet, func(path string, d fs.DirEntry) error {
		result++
		return nil
	})
	return result, err
}

func (s *Set) SIsMember(ctx context.Context, member string) (bool, error) {
	_, found, err := s.Base.CheckReadKVDataFile(s.Base.Md5(member), internal.DataTypeSet, false)
	return found, err
}

func (s *Set) SMIsMember(ctx context.Context, members []string) ([]bool, error) {
	result := make([]bool, len(members))
	for i, member := range members {
		_, found, err := s.Base.CheckReadKVDataFile(s.Base.Md5(member), internal.DataTypeSet, false)
		if err != nil {
			return nil, err
		}
		result[i] = found
	}
	return result, nil
}

func (s *Set) SPop(ctx context.Context) (v string, found bool, err error) {
	var lastErr error
	err = s.SRange(ctx, func(val string) bool {
		lastErr = s.SRem(ctx, val)
		if lastErr == nil {
			v = val
			found = true
			return false
		}
		return true
	})
	if err != nil {
		return "", false, err
	}
	if lastErr != nil {
		return "", false, lastErr
	}
	return v, found, nil
}

func (s *Set) SPopN(ctx context.Context, count int) (members []string, err error) {
	var lastErr error
	err = s.SRange(ctx, func(val string) bool {
		lastErr = s.SRem(ctx, val)
		if lastErr == nil {
			members = append(members, val)
		}
		return len(members) < count
	})
	if err != nil {
		return members, err
	}
	return members, lastErr
}

func (s *Set) SRandMember(ctx context.Context) (member string, found bool, err error) {
	err = s.SRange(ctx, func(val string) bool {
		member = val
		found = true
		return false
	})
	return member, found, err
}

func (s *Set) SRandMemberN(ctx context.Context, count int) (members []string, err error) {
	err = s.SRange(ctx, func(val string) bool {
		members = append(members, val)
		return len(members) < count
	})
	return members, err
}
