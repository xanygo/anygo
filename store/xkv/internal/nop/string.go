package nop

import (
	"context"
)

type String[V any] struct{}

func (n String[V]) Set(ctx context.Context, value V) error {
	return nil
}

func (n String[V]) Get(ctx context.Context) (v V, found bool, err error) {
	return v, false, nil
}

func (n String[V]) Incr(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n String[V]) Decr(ctx context.Context) (int64, error) {
	return 0, nil
}
