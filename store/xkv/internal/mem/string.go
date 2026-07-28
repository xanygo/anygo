package mem

import (
	"context"
	"strconv"

	"github.com/xanygo/anygo/store/xkv/internal"
)

type String struct {
	Base *Base
	Key  string
}

func (m *String) Set(ctx context.Context, value string) error {
	m.Base.setLocked(m.Key, value, internal.DataTypeString)
	return nil
}

func (m *String) Get(ctx context.Context) (string, bool, error) {
	value, found, err := m.Base.getLocked(m.Key, internal.DataTypeString)
	str, _ := value.(string)
	return str, found, err
}

func (m *String) GetDel(ctx context.Context) (value string, found bool, err error) {
	err = m.withWrite(func(val string, has bool) error {
		if !has {
			return nil
		}
		found = true
		value = val
		m.Base.deleteNoLock(m.Key)
		return nil
	})
	return value, found, err
}

func (m *String) withWrite(fn func(value string, found bool) error) error {
	return m.Base.withLock(func() error {
		value, found := m.Base.values[m.Key]
		if !found {
			return fn("", found)
		}
		typ := m.Base.keyTypes[m.Key]
		if typ != internal.DataTypeString {
			return internal.ErrInvalidType
		}
		return fn(value.(string), found)
	})
}

func (m *String) Incr(ctx context.Context) (result int64, err error) {
	return m.IncrBy(ctx, 1)
}

func (m *String) IncrBy(ctx context.Context, incr int64) (result int64, err error) {
	err = m.withWrite(func(value string, found bool) error {
		if !found {
			result = incr
			m.Base.values[m.Key] = strconv.FormatInt(result, 10)
			m.Base.keyTypes[m.Key] = internal.DataTypeString
			return nil
		}
		old, err1 := strconv.ParseInt(value, 10, 64)
		if err1 != nil {
			return err1
		}
		result = old + incr
		m.Base.values[m.Key] = strconv.FormatInt(result, 10)
		return nil
	})
	return result, err
}

func (m *String) IncrByFloat(ctx context.Context, incr float64) (result float64, err error) {
	err = m.withWrite(func(value string, found bool) error {
		if !found {
			result = incr
			m.Base.values[m.Key] = strconv.FormatFloat(result, 'g', -1, 64)
			m.Base.keyTypes[m.Key] = internal.DataTypeString
			return nil
		}
		old, err1 := strconv.ParseFloat(value, 64)
		if err1 != nil {
			return err1
		}
		result = old + incr
		m.Base.values[m.Key] = strconv.FormatFloat(result, 'g', -1, 64)
		return nil
	})
	return result, err
}

func (m *String) Decr(ctx context.Context) (result int64, err error) {
	return m.IncrBy(ctx, -1)
}
