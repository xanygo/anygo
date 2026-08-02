package rds

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xredis"
)

type Hash struct {
	Client *xredis.Client
	Key    string
}

func (kvh *Hash) HSet(ctx context.Context, field string, value string) error {
	_, err := kvh.Client.HSet(ctx, kvh.Key, field, value)
	return err
}

func (kvh *Hash) HMSet(ctx context.Context, values map[string]string) error {
	return kvh.Client.HMSet(ctx, kvh.Key, values)
}

func (kvh *Hash) HGet(ctx context.Context, field string) (string, bool, error) {
	value, err := kvh.Client.HGet(ctx, kvh.Key, field)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (kvh *Hash) HMGet(ctx context.Context, fields ...string) (map[string]string, error) {
	value, err := kvh.Client.HMGet(ctx, kvh.Key, fields...)
	if errors.Is(err, xredis.ErrNil) {
		return nil, nil
	}
	return value, nil
}

func (kvh *Hash) HDel(ctx context.Context, fields ...string) error {
	_, err := kvh.Client.HDel(ctx, kvh.Key, fields...)
	return err
}

func (kvh *Hash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	// todo: scan
	values, err := kvh.HGetAll(ctx)
	if err != nil {
		return err
	}
	for k, v := range values {
		if !fn(k, v) {
			return nil
		}
	}
	return nil
}

func (kvh *Hash) HGetAll(ctx context.Context) (map[string]string, error) {
	return kvh.Client.HGetAll(ctx, kvh.Key)
}

func (kvh *Hash) HExists(ctx context.Context, field string) (bool, error) {
	return kvh.Client.HExists(ctx, kvh.Key, field)
}

func (kvh *Hash) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	return kvh.Client.HIncrBy(ctx, kvh.Key, field, increment)
}

func (kvh *Hash) HLen(ctx context.Context) (int64, error) {
	return kvh.Client.HLen(ctx, kvh.Key)
}
