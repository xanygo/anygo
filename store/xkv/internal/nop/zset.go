package nop

import "context"

type ZSet[V any] struct{}

func (n ZSet[V]) ZAdd(ctx context.Context, score float64, member V) error {
	return nil
}

func (n ZSet[V]) ZScore(ctx context.Context, member V) (s float64, ok bool, err error) {
	return 0, false, nil
}

func (n ZSet[V]) ZRange(ctx context.Context, fn func(member V, score float64) bool) error {
	return nil
}

func (n ZSet[V]) ZIncrBy(ctx context.Context, score float64, member V) (float64, error) {
	return 0, nil
}

func (n ZSet[V]) ZRem(ctx context.Context, members ...V) error {
	return nil
}
