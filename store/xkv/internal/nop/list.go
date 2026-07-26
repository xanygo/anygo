package nop

import (
	"context"
)

type List[V any] struct{}

func (n List[V]) LPush(ctx context.Context, values ...V) (int64, error) {
	return 0, nil
}

func (n List[V]) RPush(ctx context.Context, values ...V) (int64, error) {
	return 0, nil
}

func (n List[V]) LPop(ctx context.Context) (v V, ok bool, err error) {
	return v, false, nil
}

func (n List[V]) RPop(ctx context.Context) (v V, ok bool, err error) {
	return v, false, nil
}

func (n List[V]) LRem(ctx context.Context, count int64, element string) (int64, error) {
	return 0, nil
}

func (n List[V]) Range(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n List[V]) LRange(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n List[V]) RRange(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n List[V]) LLen(ctx context.Context) (int64, error) {
	return 0, nil
}
