package mem

import (
	"context"
	"slices"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type Set struct {
	Base *Base
	Key  string
}

func (m *Set) withLocked(fn func([]string) ([]string, operate, error)) error {
	return withLocked[[]string](m.Base, m.Key, internal.DataTypeZSet, fn, strSliceEmpty)
}

func (m *Set) SAdd(ctx context.Context, members ...string) (int64, error) {
	var added int64
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		var op operate
		for _, member := range members {
			if slices.Contains(list, member) {
				continue
			}
			op = opWrite
			added++
			list = append(list, member)
		}
		return list, op, nil
	})
	return added, err
}

func (m *Set) SRem(ctx context.Context, members ...string) error {
	return m.withLocked(func(list []string) ([]string, operate, error) {
		var op operate
		for _, member := range members {
			if !slices.Contains(list, member) {
				continue
			}
			list = xslice.DeleteValue(list, member)
			op = opWrite
		}
		return list, op, nil
	})
}

func (m *Set) SRange(ctx context.Context, fn func(val string) bool) error {
	list, err := m.SMembers(ctx)
	if err != nil {
		return err
	}
	for _, val := range list {
		if !fn(val) {
			return nil
		}
	}
	return nil
}

func (m *Set) SMembers(ctx context.Context) ([]string, error) {
	var list []string
	err := m.withLocked(func(values []string) ([]string, operate, error) {
		list = slices.Clone(values)
		return list, opSkip, nil
	})
	return list, err
}

func (m *Set) SCard(ctx context.Context) (int64, error) {
	var num int64
	err := m.withLocked(func(values []string) ([]string, operate, error) {
		num = int64(len(values))
		return values, opSkip, nil
	})
	return num, err
}

func (m *Set) SIsMember(ctx context.Context, member string) (ok bool, err error) {
	err = m.withLocked(func(values []string) ([]string, operate, error) {
		ok = slices.Contains(values, member)
		return values, opSkip, nil
	})
	return ok, err
}

func (m *Set) SMIsMember(ctx context.Context, members []string) (ok []bool, err error) {
	ok = make([]bool, len(members))
	err = m.withLocked(func(values []string) ([]string, operate, error) {
		for i := range members {
			ok[i] = slices.Contains(values, members[i])
		}
		return values, opSkip, nil
	})
	return ok, err
}
