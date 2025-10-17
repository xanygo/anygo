//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package xkv

import "context"

var _ Storage[any] = (*NopStorage[any])(nil)

// NopStorage 一个黑洞存储实现
type NopStorage[V any] struct{}

func (n NopStorage[V]) String(key string) String[V] {
	return nopString[V]{}
}

var _ String[any] = (*nopString[any])(nil)

type nopString[V any] struct{}

func (n nopString[V]) Set(ctx context.Context, value V) error {
	return nil
}

func (n nopString[V]) Get(ctx context.Context) (v V, found bool, err error) {
	return v, false, nil
}

func (n nopString[V]) Incr(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n nopString[V]) Decr(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n NopStorage[V]) List(key string) List[V] {
	return nopList[V]{}
}

var _ List[any] = (*nopList[any])(nil)

type nopList[V any] struct{}

func (n nopList[V]) LPush(ctx context.Context, values ...V) (int64, error) {
	return 0, nil
}

func (n nopList[V]) RPush(ctx context.Context, values ...V) (int64, error) {
	return 0, nil
}

func (n nopList[V]) LPop(ctx context.Context) (v V, ok bool, err error) {
	return v, false, nil
}

func (n nopList[V]) RPop(ctx context.Context) (v V, ok bool, err error) {
	return v, false, nil
}

func (n nopList[V]) LRem(ctx context.Context, count int64, element string) (int64, error) {
	return 0, nil
}

func (n nopList[V]) Range(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n nopList[V]) LRange(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n nopList[V]) RRange(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n nopList[V]) LLen(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n NopStorage[V]) Hash(key string) Hash[V] {
	return nopHash[V]{}
}

var _ Hash[any] = (*nopHash[any])(nil)

type nopHash[V any] struct{}

func (n nopHash[V]) HSet(ctx context.Context, field string, value V) error {
	return nil
}

func (n nopHash[V]) HMSet(ctx context.Context, data map[string]V) error {
	return nil
}

func (n nopHash[V]) HGet(ctx context.Context, field string) (v V, ok bool, err error) {
	return v, false, nil
}

func (n nopHash[V]) HDel(ctx context.Context, fields ...string) error {
	return nil
}

func (n nopHash[V]) HRange(ctx context.Context, fn func(field string, value V) bool) error {
	return nil
}

func (n nopHash[V]) HGetAll(ctx context.Context) (map[string]V, error) {
	return nil, nil
}

func (n NopStorage[V]) Set(key string) Set[V] {
	return nopSet[V]{}
}

var _ Set[any] = (*nopSet[any])(nil)

type nopSet[V any] struct{}

func (n nopSet[V]) SAdd(ctx context.Context, members ...V) (int64, error) {
	return 0, nil
}

func (n nopSet[V]) SRem(ctx context.Context, members ...V) error {
	return nil
}

func (n nopSet[V]) SRange(ctx context.Context, fn func(val V) bool) error {
	return nil
}

func (n nopSet[V]) SMembers(ctx context.Context) ([]V, error) {
	return nil, nil
}

func (n nopSet[V]) SCard(ctx context.Context) (int64, error) {
	return 0, nil
}

func (n NopStorage[V]) ZSet(key string) ZSet[V] {
	return nopZSet[V]{}
}

var _ ZSet[any] = (*nopZSet[any])(nil)

type nopZSet[V any] struct{}

func (n nopZSet[V]) ZAdd(ctx context.Context, score float64, member V) error {
	return nil
}

func (n nopZSet[V]) ZScore(ctx context.Context, member V) (s float64, ok bool, err error) {
	return 0, false, nil
}

func (n nopZSet[V]) ZRange(ctx context.Context, fn func(member V, score float64) bool) error {
	return nil
}

func (n nopZSet[V]) ZRem(ctx context.Context, members ...V) error {
	return nil
}

func (n NopStorage[V]) Delete(ctx context.Context, keys ...string) error {
	return nil
}
