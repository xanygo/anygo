package rds

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xredis"
)

type Set struct {
	Client *xredis.Client
	Key    string
}

func (kvs *Set) SAdd(ctx context.Context, members ...string) (int64, error) {
	return kvs.Client.SAdd(ctx, kvs.Key, members...)
}

func (kvs *Set) SRem(ctx context.Context, members ...string) error {
	_, err := kvs.Client.SRem(ctx, kvs.Key, members...)
	return err
}

func (kvs *Set) SRange(ctx context.Context, fn func(val string) bool) error {
	// todo: sscan
	values, err := kvs.SMembers(ctx)
	if err != nil {
		return nil
	}
	for _, val := range values {
		if !fn(val) {
			return nil
		}
	}
	return nil
}

func (kvs *Set) SMembers(ctx context.Context) ([]string, error) {
	values, err := kvs.Client.SMembers(ctx, kvs.Key)
	if errors.Is(err, xredis.ErrNil) {
		return nil, nil
	}
	return values, err
}

func (kvs *Set) SIsMember(ctx context.Context, member string) (bool, error) {
	return kvs.Client.SIsMember(ctx, kvs.Key, member)
}

func (kvs *Set) SMIsMember(ctx context.Context, members []string) ([]bool, error) {
	return kvs.Client.SMIsMember(ctx, kvs.Key, members...)
}

func (kvs *Set) SCard(ctx context.Context) (int64, error) {
	return kvs.Client.SCard(ctx, kvs.Key)
}

func (kvs *Set) SPop(ctx context.Context) (string, bool, error) {
	return kvs.Client.SPop(ctx, kvs.Key)
}

func (kvs *Set) SPopN(ctx context.Context, count int) ([]string, error) {
	return kvs.Client.SPopN(ctx, kvs.Key, count)
}

func (kvs *Set) SRandMember(ctx context.Context) (string, bool, error) {
	result, err := kvs.Client.SRandMember(ctx, kvs.Key, 1)
	if err != nil || len(result) != 1 {
		return "", false, err
	}
	return result[0], true, nil
}

func (kvs *Set) SRandMemberN(ctx context.Context, count int) ([]string, error) {
	return kvs.Client.SRandMember(ctx, kvs.Key, count)
}
