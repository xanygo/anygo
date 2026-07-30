package file

import (
	"context"
	"io/fs"
	"os"
	"slices"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/safely"
)

type Set struct {
	Compact func()
	Base    *Base
}

func (s *Set) saveMember(member string) (added bool, err error) {
	return s.Base.writeMemberFile2(s.Base.md5(member), member)
}

func (s *Set) deleteMember(member string) (err error) {
	return s.Base.deleteMemberFile(s.Base.md5(member))
}

func (s *Set) deleteMemberByPath(fp string) (err error) {
	return s.Base.osRemove(fp)
}

func (s *Set) readMemberByPath(path string) (member string, err error) {
	bf, err1 := os.ReadFile(path)
	return string(bf), err1
}

func (s *Set) hasMember(member string) (bool, error) {
	_, found, err := s.Base.readMemberFile(s.Base.md5(member))
	return found, err
}

func (s *Set) SAdd(ctx context.Context, members ...string) (num int64, err error) {
	if len(members) == 0 {
		return 0, nil
	}
	err = s.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		for _, member := range members {
			addNew, err1 := s.saveMember(member)
			if err1 != nil {
				if num == 0 {
					s.Base.deleteKeyWhenNoMember(ctx)
				}
				return err1
			}
			if addNew {
				num++
			}
		}
		return nil
	})
	return num, err
}

func (s *Set) SRem(ctx context.Context, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	defer func() {
		go safely.RunVoid(s.Compact)
	}()
	return s.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		defer s.Base.deleteKeyWhenNoMember(ctx)
		for _, member := range members {
			if err1 := s.deleteMember(member); err1 != nil {
				return err1
			}
		}
		return nil
	})
}

// SRange 返回结果是无序的（没有按照写入顺序排序）
func (s *Set) SRange(ctx context.Context, fn func(val string) bool) error {
	return s.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return s.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			member, err1 := s.readMemberByPath(path)
			if err1 != nil {
				return err1
			}
			if !fn(member) {
				return fs.SkipAll
			}
			return nil
		})
	})
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
	err := s.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return s.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			member, err1 := s.readMemberByPath(path)
			if err1 != nil {
				return err1
			}
			info, err2 := d.Info()
			if err2 != nil {
				return err2
			}
			list = append(list, memberWithMeta{
				Member: member,
				Mtime:  info.ModTime().UnixNano(),
			})
			return nil
		})
	})

	var result []string
	slices.SortFunc(list, memberSortFn)
	for _, m := range list {
		result = append(result, m.Member)
	}
	return result, err
}

func (s *Set) SCard(ctx context.Context) (num int64, err error) {
	err = s.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return s.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			num++
			return nil
		})
	})
	return num, err
}

func (s *Set) SIsMember(ctx context.Context, member string) (ok bool, err error) {
	err = s.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		found, err1 := s.hasMember(member)
		ok = found
		return err1
	})
	return ok, err
}

func (s *Set) SMIsMember(ctx context.Context, members []string) ([]bool, error) {
	result := make([]bool, len(members))
	if len(members) == 0 {
		return result, nil
	}
	err := s.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		for idx, member := range members {
			found, err1 := s.hasMember(member)
			if err1 != nil {
				return err1
			}
			result[idx] = found
		}
		return nil
	})
	return result, err
}

func (s *Set) SPop(ctx context.Context) (v string, found bool, err error) {
	err = s.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		defer s.Base.deleteKeyWhenNoMember(ctx)
		return s.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			member, err1 := s.readMemberByPath(path)
			if err1 != nil {
				return err1
			}
			if err2 := s.deleteMemberByPath(path); err2 != nil {
				return err2
			}
			v = member
			found = true

			return fs.SkipAll
		})
	})
	if err != nil {
		return "", false, err
	}
	return v, found, nil
}

func (s *Set) SPopN(ctx context.Context, count int) (members []string, err error) {
	if count <= 0 {
		return nil, nil
	}
	err = s.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		defer s.Base.deleteKeyWhenNoMember(ctx)
		return s.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			member, err1 := s.readMemberByPath(path)
			if err1 != nil {
				return err1
			}
			if err2 := s.deleteMemberByPath(path); err2 != nil {
				return err2
			}
			members = append(members, member)
			if len(members) == count {
				return fs.SkipAll
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return members, nil
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
