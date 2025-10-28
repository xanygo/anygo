//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package xkv

import (
	"context"
	"errors"

	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/xcodec"
)

var _ Storage[any] = (*Transformer[any])(nil)

// Transformer 将 StringStorage  转换为值可以存储任意类型的
type Transformer[V any] struct {
	Storage StringStorage
	Codec   xcodec.Codec
}

func (tr Transformer[V]) String(key string) String[V] {
	return AsString[V](tr.Storage, tr.Codec, key)
}

func AsString[V any](s StringStorage, codec xcodec.Codec, key string) String[V] {
	return transString[V]{
		ss:    s.String(key),
		codec: codec,
	}
}

var _ String[any] = (*transString[any])(nil)

type transString[V any] struct {
	ss    String[string]
	codec xcodec.Codec
}

func (ts transString[V]) Set(ctx context.Context, value V) error {
	str, err := xcodec.EncodeToString(ts.codec, value)
	if err != nil {
		return err
	}
	return ts.ss.Set(ctx, str)
}

func (ts transString[V]) Get(ctx context.Context) (v V, found bool, err error) {
	str, found, err := ts.ss.Get(ctx)
	if err != nil {
		return v, false, err
	}
	if !found {
		return v, false, nil
	}
	err = xcodec.DecodeFromString(ts.codec, str, &v)
	if err != nil {
		return v, false, err
	}
	return v, true, err
}

func (ts transString[V]) Incr(ctx context.Context) (int64, error) {
	return ts.ss.Incr(ctx)
}

func (ts transString[V]) Decr(ctx context.Context) (int64, error) {
	return ts.ss.Decr(ctx)
}

func (tr Transformer[V]) List(key string) List[V] {
	return AsList[V](tr.Storage, tr.Codec, key)
}

func AsList[V any](s StringStorage, codec xcodec.Codec, key string) List[V] {
	return transList[V]{
		ss:    s.List(key),
		codec: codec,
	}
}

var _ List[any] = (*transList[any])(nil)

type transList[V any] struct {
	ss    List[string]
	codec xcodec.Codec
}

func (t transList[V]) LPush(ctx context.Context, values ...V) (int64, error) {
	ms, errs := internal.EncodeToStrings(t.codec, values)
	if len(errs) == len(values) {
		return 0, errors.Join(errs...)
	}
	num, err := t.ss.LPush(ctx, ms...)
	if err != nil {
		errs = append(errs, err)
	}
	return num, errors.Join(errs...)
}

func (t transList[V]) RPush(ctx context.Context, values ...V) (int64, error) {
	ms, errs := internal.EncodeToStrings(t.codec, values)
	if len(errs) == len(values) {
		return 0, errors.Join(errs...)
	}
	num, err := t.ss.RPush(ctx, ms...)
	if err != nil {
		errs = append(errs, err)
	}
	return num, errors.Join(errs...)
}

func (t transList[V]) LPop(ctx context.Context) (v V, ok bool, err error) {
	str, found, err := t.ss.LPop(ctx)
	if !found || err != nil {
		return v, false, err
	}
	err = xcodec.DecodeFromString(t.codec, str, &v)
	return v, err == nil, err
}

func (t transList[V]) RPop(ctx context.Context) (v V, ok bool, err error) {
	str, found, err := t.ss.RPop(ctx)
	if !found || err != nil {
		return v, false, err
	}
	err = xcodec.DecodeFromString(t.codec, str, &v)
	return v, err == nil, err
}

func (t transList[V]) LRem(ctx context.Context, count int64, element string) (int64, error) {
	return t.ss.LRem(ctx, count, element)
}

func (t transList[V]) Range(ctx context.Context, fn func(val V) bool) error {
	var decodeErr error
	err := t.ss.Range(ctx, func(val string) bool {
		var v V
		decodeErr = xcodec.DecodeFromString(t.codec, val, &v)
		if decodeErr != nil {
			return false
		}
		return fn(v)
	})
	if decodeErr != nil {
		return decodeErr
	}
	return err
}

func (t transList[V]) LRange(ctx context.Context, fn func(val V) bool) error {
	var decodeErr error
	err := t.ss.LRange(ctx, func(val string) bool {
		var v V
		decodeErr = xcodec.DecodeFromString(t.codec, val, &v)
		if decodeErr != nil {
			return false
		}
		return fn(v)
	})
	if decodeErr != nil {
		return decodeErr
	}
	return err
}

func (t transList[V]) RRange(ctx context.Context, fn func(val V) bool) error {
	var decodeErr error
	err := t.ss.RRange(ctx, func(val string) bool {
		var v V
		decodeErr = xcodec.DecodeFromString(t.codec, val, &v)
		if decodeErr != nil {
			return false
		}
		return fn(v)
	})
	if decodeErr != nil {
		return decodeErr
	}
	return err
}

func (t transList[V]) LLen(ctx context.Context) (int64, error) {
	return t.ss.LLen(ctx)
}

func (tr Transformer[V]) Hash(key string) Hash[V] {
	return AsHash[V](tr.Storage, tr.Codec, key)
}

func AsHash[V any](s StringStorage, codec xcodec.Codec, key string) Hash[V] {
	return transHash[V]{
		ss:    s.Hash(key),
		codec: codec,
	}
}

var _ Hash[any] = (*transHash[any])(nil)

type transHash[V any] struct {
	ss    Hash[string]
	codec xcodec.Codec
}

func (t transHash[V]) HSet(ctx context.Context, field string, value V) error {
	str, err := xcodec.EncodeToString(t.codec, value)
	if err != nil {
		return err
	}
	return t.ss.HSet(ctx, field, str)
}

func (t transHash[V]) HMSet(ctx context.Context, data map[string]V) error {
	mp, errs := internal.EncodeMapValueToStrings(t.codec, data)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return t.ss.HMSet(ctx, mp)
}

func (t transHash[V]) HGet(ctx context.Context, field string) (v V, ok bool, err error) {
	str, found, err := t.ss.HGet(ctx, field)
	if !found || err != nil {
		return v, false, err
	}
	err = xcodec.DecodeFromString(t.codec, str, &v)
	return v, err == nil, err
}

func (t transHash[V]) HDel(ctx context.Context, fields ...string) error {
	return t.ss.HDel(ctx, fields...)
}

func (t transHash[V]) HRange(ctx context.Context, fn func(field string, value V) bool) error {
	var decodeErr error
	err := t.ss.HRange(ctx, func(field string, value string) bool {
		var v V
		decodeErr = xcodec.DecodeFromString(t.codec, value, &v)
		if decodeErr != nil {
			return false
		}
		return fn(field, v)
	})
	if decodeErr != nil {
		return decodeErr
	}
	return err
}

func (t transHash[V]) HGetAll(ctx context.Context) (map[string]V, error) {
	result := make(map[string]V)
	err := t.HRange(ctx, func(field string, value V) bool {
		result[field] = value
		return true
	})
	return result, err
}

func (tr Transformer[V]) Set(key string) Set[V] {
	return AsSet[V](tr.Storage, tr.Codec, key)
}

func AsSet[V any](s StringStorage, codec xcodec.Codec, key string) Set[V] {
	return transSet[V]{
		ss:    s.Set(key),
		codec: codec,
	}
}

var _ Set[any] = (*transSet[any])(nil)

type transSet[V any] struct {
	ss    Set[string]
	codec xcodec.Codec
}

func (t transSet[V]) SAdd(ctx context.Context, members ...V) (int64, error) {
	ms, errs := internal.EncodeToStrings(t.codec, members)
	if len(errs) == len(members) {
		return 0, errors.Join(errs...)
	}
	return t.ss.SAdd(ctx, ms...)
}

func (t transSet[V]) SRem(ctx context.Context, members ...V) error {
	ms, errs := internal.EncodeToStrings(t.codec, members)
	if len(errs) == len(members) {
		return errors.Join(errs...)
	}
	if err := t.ss.SRem(ctx, ms...); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (t transSet[V]) SRange(ctx context.Context, fn func(val V) bool) error {
	var decodeErr error
	err := t.ss.SRange(ctx, func(value string) bool {
		var v V
		decodeErr = xcodec.DecodeFromString(t.codec, value, &v)
		if decodeErr != nil {
			return false
		}
		return fn(v)
	})

	if decodeErr != nil {
		return decodeErr
	}
	return err
}

func (t transSet[V]) SMembers(ctx context.Context) ([]V, error) {
	var result []V
	err := t.SRange(ctx, func(val V) bool {
		result = append(result, val)
		return true
	})
	return result, err
}

func (t transSet[V]) SCard(ctx context.Context) (int64, error) {
	return t.ss.SCard(ctx)
}

func (tr Transformer[V]) ZSet(key string) ZSet[V] {
	return AsZSet[V](tr.Storage, tr.Codec, key)
}

func AsZSet[V any](s StringStorage, codec xcodec.Codec, key string) ZSet[V] {
	return transZSet[V]{
		ss:    s.ZSet(key),
		codec: codec,
	}
}

var _ ZSet[any] = (*transZSet[any])(nil)

type transZSet[V any] struct {
	ss    ZSet[string]
	codec xcodec.Codec
}

func (t transZSet[V]) ZAdd(ctx context.Context, score float64, member V) error {
	str, err := xcodec.EncodeToString(t.codec, member)
	if err != nil {
		return err
	}
	return t.ss.ZAdd(ctx, score, str)
}

func (t transZSet[V]) ZScore(ctx context.Context, member V) (float64, bool, error) {
	str, err := xcodec.EncodeToString(t.codec, member)
	if err != nil {
		return 0, false, err
	}
	return t.ss.ZScore(ctx, str)
}

func (t transZSet[V]) ZRange(ctx context.Context, fn func(member V, score float64) bool) error {
	var decodeErr error
	err := t.ss.ZRange(ctx, func(member string, score float64) bool {
		var v V
		decodeErr = xcodec.DecodeFromString(t.codec, member, &v)
		if decodeErr != nil {
			return false
		}
		return fn(v, score)
	})
	if decodeErr != nil {
		return decodeErr
	}
	return err
}

func (t transZSet[V]) ZRem(ctx context.Context, members ...V) error {
	ms, errs := internal.EncodeToStrings(t.codec, members)
	if len(errs) == len(members) {
		return errors.Join(errs...)
	}
	return t.ss.ZRem(ctx, ms...)
}

func (tr Transformer[V]) Delete(ctx context.Context, keys ...string) error {
	return tr.Storage.Delete(ctx, keys...)
}

func (tr Transformer[V]) Has(ctx context.Context, key string) (bool, error) {
	return tr.Storage.Has(ctx, key)
}
