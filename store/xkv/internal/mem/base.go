package mem

import (
	"context"
	"sync"

	"github.com/xanygo/anygo/store/xkv/internal"
)

func NewBase() *Base {
	return &Base{
		values:   make(map[string]any),
		keyTypes: make(map[string]internal.DataType),
	}
}

type Base struct {
	values   map[string]any
	keyTypes map[string]internal.DataType
	mux      sync.RWMutex
}

func (m *Base) deleteNoLock(key string) {
	delete(m.values, key)
	delete(m.keyTypes, key)
}

func (m *Base) getLocked(key string, wantType internal.DataType) (value any, found bool, err error) {
	var tp internal.DataType
	m.mux.RLock()
	value, found = m.values[key]
	tp = m.keyTypes[key]
	m.mux.RUnlock()
	if found && tp != wantType {
		return "", false, internal.ErrInvalidType
	}
	return value, found, nil
}

func (m *Base) withLock(fn func() error) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	return fn()
}

func (m *Base) Delete(ctx context.Context, keys ...string) error {
	return m.withLock(func() error {
		for _, key := range keys {
			delete(m.values, key)
			delete(m.keyTypes, key)
		}
		return nil
	})
}

func (m *Base) Has(ctx context.Context, key string) (found bool, err error) {
	err = m.withLock(func() error {
		_, found = m.values[key]
		return nil
	})
	return found, err
}

type operate uint8

const (
	opSkip  operate = 0
	opWrite operate = 1
)

func withLocked[T any](
	base *Base,
	key string,
	dt internal.DataType,
	fn func(T) (T, operate, error),
	dataEmpty func(T) bool,
) error {
	return base.withLock(func() error {
		value, found := base.values[key]
		var result T
		var op operate
		var err error
		if !found {
			var zero T
			result, op, err = fn(zero)
		} else {
			tp := base.keyTypes[key]
			if tp != dt {
				return internal.ErrInvalidType
			}
			result, op, err = fn(value.(T))
		}
		if err != nil {
			return err
		}

		if op == opWrite {
			if dataEmpty(result) {
				delete(base.values, key)
				delete(base.keyTypes, key)
			} else {
				base.values[key] = result
				base.keyTypes[key] = dt
			}
		}
		return nil
	})
}
