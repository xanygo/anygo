//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xcache

import (
	"context"
	"errors"
	"time"

	"github.com/xanygo/anygo/xcodec"
)

var _ Cache[string, any] = (*Transformer[any])(nil)
var _ MCache[string, any] = (*Transformer[any])(nil)

// Transformer 使用底层存储 K-V 均为 string 类型的 cache，存储缓存数据
type Transformer[V any] struct {
	Cache StringCache
	Codec xcodec.Codec
}

func (t *Transformer[V]) Get(ctx context.Context, key string) (value V, err error) {
	str, err := t.Cache.Get(ctx, key)
	if err != nil {
		return value, err
	}
	err = xcodec.DecodeFromString(t.Codec, str, &value)
	return value, err
}

func (t *Transformer[V]) MGet(ctx context.Context, keys ...string) (result map[string]V, err error) {
	if mc, ok := t.Cache.(MGetter[string, string]); ok {
		rt, err1 := mc.MGet(ctx, keys...)
		if len(rt) == 0 {
			return result, err1
		}
		var errs []error
		if err1 != nil {
			errs = append(errs, err1)
		}
		result = make(map[string]V, len(keys))
		for key, strVal := range rt {
			var value V
			err2 := xcodec.DecodeFromString(t.Codec, strVal, &value)
			if err2 == nil {
				result[key] = value
			} else {
				errs = append(errs, err2)
			}
		}
		return result, errors.Join(errs...)
	}
	var errs []error
	result = make(map[string]V, len(keys))
	for _, key := range keys {
		select {
		case <-ctx.Done():
			return result, context.Cause(ctx)
		default:
		}
		value, err3 := t.Get(ctx, key)
		if err3 == nil {
			result[key] = value
		} else {
			errs = append(errs, err3)
		}
	}
	return result, errors.Join(errs...)
}

func (t *Transformer[V]) Set(ctx context.Context, key string, value V, ttl time.Duration) error {
	str, err := xcodec.EncodeToString(t.Codec, value)
	if err != nil {
		return err
	}
	return t.Cache.Set(ctx, key, str, ttl)
}

func (t *Transformer[V]) MSet(ctx context.Context, values map[string]V, ttl time.Duration) error {
	if mc, ok := t.Cache.(MSetter[string, string]); ok {
		var errs []error
		kv := make(map[string]string, len(values))
		for key, value := range values {
			str, err := xcodec.EncodeToString(t.Codec, value)
			if err == nil {
				kv[key] = str
			} else {
				errs = append(errs, err)
			}
		}
		if len(kv) == 0 {
			return errors.Join(errs...)
		}
		if err := mc.MSet(ctx, kv, ttl); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}

	var errs []error
	for key, value := range values {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		if err := t.Set(ctx, key, value, ttl); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (t *Transformer[V]) Delete(ctx context.Context, keys ...string) error {
	return t.Cache.Delete(ctx, keys...)
}

var _ HasStats = (*Transformer[any])(nil)

func (t *Transformer[V]) Stats() Stats {
	if hs, ok := t.Cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{}
}

var _ HasAllStats = (*Transformer[any])(nil)

func (t *Transformer[V]) AllStats() map[string]Stats {
	if hs, ok := t.Cache.(HasAllStats); ok {
		return hs.AllStats()
	}
	return nil
}
