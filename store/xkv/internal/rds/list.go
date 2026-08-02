package rds

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xredis"
)

var _ xkv.List[string] = (*List)(nil)

type List struct {
	Client *xredis.Client
	Key    string
}

func (l *List) LPush(ctx context.Context, values ...string) (int64, error) {
	return l.Client.LPush(ctx, l.Key, values...)
}

func (l *List) RPush(ctx context.Context, values ...string) (int64, error) {
	return l.Client.RPush(ctx, l.Key, values...)
}

func (l *List) LPop(ctx context.Context) (string, bool, error) {
	value, err := l.Client.LPop(ctx, l.Key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (l *List) LPopN(ctx context.Context, count int) ([]string, error) {
	values, err := l.Client.LPopN(ctx, l.Key, int64(count))
	if errors.Is(err, xredis.ErrNil) {
		return nil, nil
	}
	return values, err
}

func (l *List) RPop(ctx context.Context) (string, bool, error) {
	value, err := l.Client.RPop(ctx, l.Key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (l *List) RPopN(ctx context.Context, count int) ([]string, error) {
	values, err := l.Client.RPopN(ctx, l.Key, int64(count))
	if errors.Is(err, xredis.ErrNil) {
		return nil, nil
	}
	return values, err
}

func (l *List) LRem(ctx context.Context, count int64, element string) (int64, error) {
	return l.Client.LRem(ctx, l.Key, count, element)
}

func (l *List) Range(ctx context.Context, fn func(val string) bool) error {
	return l.LRange(ctx, fn)
}

func (l *List) LRange(ctx context.Context, fn func(val string) bool) error {
	for start := int64(0); ; start += 10 {
		stop := start + 10
		values, err := l.Client.LRange(ctx, l.Key, start, stop)
		if errors.Is(err, xredis.ErrNil) || len(values) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		for _, val := range values {
			if !fn(val) {
				return nil
			}
		}
	}
}

func (l *List) RRange(ctx context.Context, fn func(val string) bool) error {
	for stop := int64(-1); ; stop -= 9 {
		start := stop - 9
		values, err := l.Client.LRange(ctx, l.Key, start, stop)
		if errors.Is(err, xredis.ErrNil) || len(values) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		for i := len(values) - 1; i >= 0; i-- {
			if !fn(values[i]) {
				return nil
			}
		}
	}
}

func (l *List) LLen(ctx context.Context) (int64, error) {
	return l.Client.LLen(ctx, l.Key)
}
