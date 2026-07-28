package nop

import (
	"context"
)

type Hash[V any] struct{}

func (n Hash[V]) HSet(ctx context.Context, field string, value V) error {
	return nil
}

func (n Hash[V]) HMSet(ctx context.Context, data map[string]V) error {
	return nil
}

func (n Hash[V]) HGet(ctx context.Context, field string) (v V, ok bool, err error) {
	return v, false, nil
}

func (n Hash[V]) HDel(ctx context.Context, fields ...string) error {
	return nil
}

func (n Hash[V]) HRange(ctx context.Context, fn func(field string, value V) bool) error {
	return nil
}

func (n Hash[V]) HGetAll(ctx context.Context) (map[string]V, error) {
	return nil, nil
}

func (n Hash[V]) HExists(ctx context.Context, field string) (bool, error) {
	return false, nil
}

func (n Hash[V]) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	return increment, nil
}
