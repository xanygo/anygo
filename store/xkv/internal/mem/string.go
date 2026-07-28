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
	err = m.withWrite(func(value string, found bool) error {
		if !found {
			result = 1
			m.Base.values[m.Key] = "1"
			m.Base.keyTypes[m.Key] = internal.DataTypeString
			return nil
		}
		old, err1 := strconv.ParseInt(value, 10, 64)
		if err1 != nil {
			return err1
		}
		result = old + 1
		m.Base.values[m.Key] = strconv.FormatInt(result, 10)
		return nil
	})
	return result, err
}

func (m *String) Decr(ctx context.Context) (result int64, err error) {
	err = m.withWrite(func(value string, found bool) error {
		if !found {
			result = -1
			m.Base.values[m.Key] = "-1"
			m.Base.keyTypes[m.Key] = internal.DataTypeString
			return nil
		}
		old, err1 := strconv.ParseInt(value, 10, 64)
		if err1 != nil {
			return err1
		}
		result = old - 1
		m.Base.values[m.Key] = strconv.FormatInt(result, 10)
		return nil
	})
	return result, err
}
