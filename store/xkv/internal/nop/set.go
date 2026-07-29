package nop

import (
	"context"
)

type Set[V any] struct{}

func (n Set[V]) SAdd(ctx context.Context, members ...V) (int64, error) {
	return 0, nil
}

func (n Set[V]) SRem(ctx context.Context, members ...V) error {
	return nil
}

func (n Set[V]) SRange(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n Set[V]) SMembers(ctx context.Context) ([]V, error) {
	return nil, nil
}

func (n Set[V]) SCard(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n Set[V]) SIsMember(ctx context.Context, member V) (bool, error) {
	return false, nil
}

func (n Set[V]) SMIsMember(ctx context.Context, members []V) ([]bool, error) {
	result := make([]bool, len(members))
	return result, nil
}
