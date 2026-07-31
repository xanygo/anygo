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

func (n ZSet[V]) ZRangeByScore(ctx context.Context, min string, max string, fn func(member V, score float64) bool) error {
	return nil
}

func (n ZSet[V]) ZIncrBy(ctx context.Context, score float64, member V) (float64, error) {
	return 0, nil
}

func (n ZSet[V]) ZCount(ctx context.Context, min, max string) (int64, error) {
	return 0, nil
}

func (n ZSet[V]) ZLen(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n ZSet[V]) ZRem(ctx context.Context, members ...V) error {
	return nil
}

func (n ZSet[V]) ZRemRangeByScore(ctx context.Context, min, max string) (int64, error) {
	return 0, nil
}

func (n ZSet[V]) ZRank(ctx context.Context, member V) (int64, float64, error) {
	return 0, 0, nil
}

func (n ZSet[V]) ZPopMax(ctx context.Context, count int) ([]V, []float64, error) {
	return nil, nil, nil
}

func (n ZSet[V]) ZPopMin(ctx context.Context, count int) ([]V, []float64, error) {
	return nil, nil, nil
}
