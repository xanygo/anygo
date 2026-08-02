package mem

import (
	"context"
	"maps"
	"strconv"

	"github.com/xanygo/anygo/store/xkv/internal"
)

type Hash struct {
	Base *Base
	Key  string
}

func strMapEmpty(d map[string]string) bool {
	return len(d) == 0
}

func (m *Hash) withLocked(fn func(map[string]string) (map[string]string, operate, error)) error {
	return withLocked[map[string]string](m.Base, m.Key, internal.DataTypeHash, func(m map[string]string) (map[string]string, operate, error) {
		if m == nil {
			m = make(map[string]string)
		}
		return fn(m)
	}, strMapEmpty)
}

func (m *Hash) HSet(ctx context.Context, field string, value string) error {
	return m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		m[field] = value
		return m, opWrite, nil
	})
}

func (m *Hash) HMSet(ctx context.Context, values map[string]string) error {
	return m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		maps.Copy(m, values)
		return m, opWrite, nil
	})
}

func (m *Hash) HGet(ctx context.Context, field string) (string, bool, error) {
	var value string
	var found bool
	err := m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		value, found = m[field]
		return m, opSkip, nil
	})
	return value, found, err
}

func (m *Hash) HMGet(ctx context.Context, fields ...string) (result map[string]string, err error) {
	if len(fields) == 0 {
		return nil, nil
	}
	result = make(map[string]string, len(fields))
	err = m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		for _, field := range fields {
			if value, found := m[field]; found {
				result[field] = value
			}
		}
		return m, opSkip, nil
	})
	return result, err
}

func (m *Hash) HDel(ctx context.Context, fields ...string) error {
	return m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		if len(m) == 0 {
			return m, opSkip, nil
		}
		var op operate
		for _, field := range fields {
			if _, found := m[field]; found {
				op = opWrite
			}
			delete(m, field)
		}
		return m, op, nil
	})
}

func (m *Hash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	values, err := m.HGetAll(ctx)
	if err != nil {
		return err
	}
	for key, value := range values {
		if !fn(key, value) {
			return nil
		}
	}
	return err
}

func (m *Hash) HGetAll(ctx context.Context) (map[string]string, error) {
	var result map[string]string
	err := m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		result = maps.Clone(m)
		return m, opSkip, nil
	})
	return result, err
}

func (m *Hash) HExists(ctx context.Context, field string) (found bool, err error) {
	err = m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		_, found = m[field]
		return m, opSkip, nil
	})
	return found, err
}

func (m *Hash) HIncrBy(ctx context.Context, field string, increment int64) (num int64, err error) {
	err = m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		value, found := m[field]
		if !found {
			num = increment
			m[field] = strconv.FormatInt(increment, 10)
		} else {
			old, err2 := strconv.ParseInt(value, 10, 64)
			if err2 != nil {
				return nil, opSkip, err2
			}
			num = old + increment
			m[field] = strconv.FormatInt(num, 10)
		}
		return m, opWrite, nil
	})
	return num, err
}

func (m *Hash) HLen(ctx context.Context) (num int64, err error) {
	err = m.withLocked(func(m map[string]string) (map[string]string, operate, error) {
		num = int64(len(m))
		return m, opSkip, nil
	})
	return num, err
}
