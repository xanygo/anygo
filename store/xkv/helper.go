package xkv

import (
	"context"
)

// ZItem zset 的一条数据
type ZItem[T any] struct {
	Member T
	Score  float64
}

type HItem[T any] struct {
	Field string
	Value T
}

func ZRange[V any](ctx context.Context, z ZSet[V], count int) ([]*ZItem[V], error) {
	var result []*ZItem[V]
	err := z.ZRange(ctx, func(item V, score float64) bool {
		result = append(result, &ZItem[V]{
			Member: item,
			Score:  score,
		})
		if count > 0 {
			return len(result) < count
		}
		return true
	})
	return result, err
}

func ZRangeByScore[V any](ctx context.Context, z ZSet[V], min, max string, count int) ([]*ZItem[V], error) {
	var result []*ZItem[V]
	err := z.ZRangeByScore(ctx, min, max, func(item V, score float64) bool {
		result = append(result, &ZItem[V]{
			Member: item,
			Score:  score,
		})
		if count > 0 {
			return len(result) < count
		}
		return true
	})
	return result, err
}

func ZPopMax[V any](ctx context.Context, z ZSet[V], count int) ([]*ZItem[V], error) {
	members, scores, err := z.ZPopMax(ctx, count)
	if err != nil {
		return nil, err
	}
	result := make([]*ZItem[V], len(members))
	for i, member := range members {
		result[i] = &ZItem[V]{
			Member: member,
			Score:  scores[i],
		}
	}
	return result, nil
}

func ZPopMin[V any](ctx context.Context, z ZSet[V], count int) ([]*ZItem[V], error) {
	members, scores, err := z.ZPopMin(ctx, count)
	if err != nil {
		return nil, err
	}
	result := make([]*ZItem[V], len(members))
	for i, member := range members {
		result[i] = &ZItem[V]{
			Member: member,
			Score:  scores[i],
		}
	}
	return result, nil
}

func SRange[V any](ctx context.Context, s Set[V], count int) ([]V, error) {
	var result []V
	err := s.SRange(ctx, func(member V) bool {
		result = append(result, member)
		if count > 0 {
			return len(result) < count
		}
		return true
	})
	return result, err
}

func HRange[V any](ctx context.Context, h Hash[V], count int) ([]*HItem[V], error) {
	var result []*HItem[V]
	err := h.HRange(ctx, func(field string, value V) bool {
		result = append(result, &HItem[V]{
			Field: field,
			Value: value,
		})
		if count > 0 {
			return len(result) < count
		}
		return true
	})
	return result, err
}

func LRange[V any](ctx context.Context, l List[V], count int) ([]V, error) {
	var result []V
	err := l.LRange(ctx, func(item V) bool {
		result = append(result, item)
		if count > 0 {
			return len(result) < count
		}
		return true
	})
	return result, err
}
