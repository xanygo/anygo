package mem

import (
	"context"
	"slices"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type List struct {
	Base *Base
	Key  string
}

func strSliceEmpty(list []string) bool {
	return len(list) == 0
}

func (m *List) withLocked(fn func([]string) ([]string, operate, error)) error {
	return withLocked[[]string](m.Base, m.Key, internal.DataTypeList, fn, strSliceEmpty)
}

func (m *List) LPush(ctx context.Context, values ...string) (int64, error) {
	var num int64
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		ret := slices.Insert(list, 0, values...)
		num = int64(len(ret))
		return ret, opWrite, nil
	})
	return num, err
}

func (m *List) RPush(ctx context.Context, values ...string) (int64, error) {
	var num int64
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		ret := append(list, values...)
		num = int64(len(ret))
		return ret, opWrite, nil
	})
	return num, err
}

func (m *List) LPop(ctx context.Context) (string, bool, error) {
	var value string
	var found bool
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		if len(list) == 0 {
			return nil, opSkip, nil
		}
		list, value, found = xslice.PopHead(list)
		return list, opWrite, nil
	})

	return value, found, err
}

func (m *List) RPop(ctx context.Context) (string, bool, error) {
	var value string
	var found bool
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		if len(list) == 0 {
			return nil, opSkip, nil
		}
		list, value, found = xslice.PopTail(list)
		return list, opWrite, nil
	})

	return value, found, err
}

func (m *List) LRem(ctx context.Context, count int64, element string) (int64, error) {
	var deleted int64
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		var op operate
		if count == 0 { // 移除所有等于 element 的元素。
			list = slices.DeleteFunc(list, func(s string) bool {
				if s == element {
					deleted++
					return true
				}
				return false
			})
			if deleted > 0 {
				op = opWrite
			}
			return list, op, nil
		} else if count > 0 {
			newList := xslice.DeleteFuncN(list, func(s string) bool {
				return s == element
			}, int(count))
			deleted = int64(len(list) - len(newList))
			if deleted > 0 {
				op = opWrite
			}
			return newList, op, nil
		} else { // if count < 0
			newList := xslice.RevDeleteFuncN(list, func(s string) bool {
				return s == element
			}, int(count*-1))
			deleted = int64(len(list) - len(newList))
			if deleted > 0 {
				op = opWrite
			}
			return newList, op, nil
		}
	})
	return deleted, err
}

func (m *List) Range(ctx context.Context, fn func(val string) bool) error {
	return m.LRange(ctx, fn)
}

func (m *List) LRange(ctx context.Context, fn func(val string) bool) error {
	var values []string
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		values = slices.Clone(list)
		return list, opSkip, nil
	})
	if err != nil {
		return err
	}
	for _, val := range values {
		if !fn(val) {
			return nil
		}
		if err = ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (m *List) RRange(ctx context.Context, fn func(val string) bool) error {
	var values []string
	err := m.withLocked(func(list []string) ([]string, operate, error) {
		values = slices.Clone(list)
		return list, opSkip, nil
	})
	if err != nil {
		return err
	}
	for i := len(values) - 1; i >= 0; i-- {
		if !fn(values[i]) {
			return nil
		}
		if err = ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (m *List) LLen(ctx context.Context) (num int64, err error) {
	err = m.withLocked(func(list []string) ([]string, operate, error) {
		num = int64(len(list))
		return list, opSkip, nil
	})
	return num, err
}
