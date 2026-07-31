package rds

import (
	"context"
	"errors"
	"io"

	"github.com/xanygo/anygo/store/xredis"
)

type ZSet struct {
	Client *xredis.Client
	Key    string
}

func (z *ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	_, err := z.Client.ZAdd(ctx, z.Key, score, member)
	return err
}

func (z *ZSet) ZIncrBy(ctx context.Context, score float64, member string) (float64, error) {
	num, err := z.Client.ZIncrBy(ctx, z.Key, score, member)
	return num, err
}

func (z *ZSet) ZCount(ctx context.Context, min, max string) (int64, error) {
	num, err := z.Client.ZCount(ctx, z.Key, min, max)
	return num, err
}

func (z *ZSet) ZLen(ctx context.Context) (int64, error) {
	num, err := z.Client.ZCount(ctx, z.Key, "-inf", "+inf")
	return num, err
}

func (z *ZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	value, err := z.Client.ZScore(ctx, z.Key, member)
	if errors.Is(err, xredis.ErrNil) {
		return 0, false, nil
	}
	return value, err == nil, err
}

func (z *ZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	return z.Client.ZScanWalk(ctx, z.Key, 0, "", 10, func(cursor uint64, items []xredis.Z) error {
		for _, item := range items {
			if !fn(item.Member, item.Score) {
				return io.EOF
			}
		}
		return nil
	})
}

func (z *ZSet) ZRangeByScore(ctx context.Context, min string, max string, fn func(member string, score float64) bool) error {
	opt := xredis.ZRangeBy{
		Start: min,
		Stop:  max,
	}
	list, err := z.Client.ZRangeByScoreWithScore(ctx, z.Key, opt)
	if err != nil {
		return nil
	}
	for _, item := range list {
		if !fn(item.Member, item.Score) {
			return nil
		}
	}
	return nil
}

func (z *ZSet) ZRank(ctx context.Context, member string) (int64, float64, error) {
	index, score, err := z.Client.ZRankWithScore(ctx, z.Key, member)
	if errors.Is(err, xredis.ErrNil) {
		return -1, score, nil
	}
	return index, score, err
}

func (z *ZSet) ZRem(ctx context.Context, members ...string) error {
	_, err := z.Client.ZRem(ctx, z.Key, members...)
	return err
}

func (z *ZSet) ZPopMax(ctx context.Context, count int) (members []string, scores []float64, err error) {
	items, err1 := z.Client.ZPopMax(ctx, z.Key, count)
	if err1 != nil {
		return nil, nil, err1
	}
	for _, item := range items {
		members = append(members, item.Member)
		scores = append(scores, item.Score)
	}
	return members, scores, nil
}

func (z *ZSet) ZPopMin(ctx context.Context, count int) (members []string, scores []float64, err error) {
	items, err1 := z.Client.ZPopMin(ctx, z.Key, count)
	if err1 != nil {
		return nil, nil, err1
	}
	for _, item := range items {
		members = append(members, item.Member)
		scores = append(scores, item.Score)
	}
	return members, scores, nil
}
