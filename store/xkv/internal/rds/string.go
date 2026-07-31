package rds

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xredis"
)

type String struct {
	Client *xredis.Client
	Key    string
}

func (s *String) Set(ctx context.Context, value string) error {
	return s.Client.Set(ctx, s.Key, value)
}

func (s *String) SetNX(ctx context.Context, value string) (bool, error) {
	return s.Client.SetNX(ctx, s.Key, value, 0)
}

func (s *String) Get(ctx context.Context) (string, bool, error) {
	value, err := s.Client.Get(ctx, s.Key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (s *String) GetSet(ctx context.Context, value string) (string, bool, error) {
	old, err := s.Client.GetSet(ctx, s.Key, value)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return old, err == nil, err
}

func (s *String) GetDel(ctx context.Context) (string, bool, error) {
	value, err := s.Client.GetDel(ctx, s.Key)
	if errors.Is(err, xredis.ErrNil) {
		return "", false, nil
	}
	return value, err == nil, err
}

func (s *String) Incr(ctx context.Context) (int64, error) {
	return s.Client.Incr(ctx, s.Key)
}

func (s *String) IncrBy(ctx context.Context, incr int64) (int64, error) {
	return s.Client.IncrBy(ctx, s.Key, incr)
}

func (s *String) IncrByFloat(ctx context.Context, incr float64) (float64, error) {
	return s.Client.IncrByFloat(ctx, s.Key, incr)
}

func (s *String) Decr(ctx context.Context) (int64, error) {
	return s.Client.Decr(ctx, s.Key)
}
