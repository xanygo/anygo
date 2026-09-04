//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xcache

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xcodec"
)

var _ Cache[string, any] = (*Transformer[string, any])(nil)
var _ MCache[string, any] = (*Transformer[string, any])(nil)

// Transformer 使用底层存储 K-V 均为 string 类型的 cache，存储缓存数据
type Transformer[k comparable, V any] struct {
	Cache      StringCache        // 必填
	KeyPrefix  string             // 可选，key 的前缀
	KeyConvert func(k any) string // 可选, 为空时，使用 zreflect.ToString
	ValueCodec xcodec.Codec       // 必填
}

func (t *Transformer[K, V]) Init(param map[string]any) error {
	if t.KeyPrefix == "" {
		t.KeyPrefix, _ = xmap.GetString(param, "KeyPrefix")
	}
	if t.ValueCodec == nil {
		name, ok := xmap.GetString(param, "ValueCodec")
		if !ok || name == "" {
			t.ValueCodec = xcodec.JSON
		} else {
			c, err := xcodec.Find(name)
			if err != nil {
				return err
			}
			t.ValueCodec = c
		}
	}
	return nil
}

func (t *Transformer[K, V]) keyAsString(key K) string {
	rv := reflect.ValueOf(key)

	switch rv.Kind() {
	case reflect.String:
		return t.KeyPrefix + rv.String()

	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return t.KeyPrefix + string(rv.Bytes())
		}
	}
	if t.KeyConvert == nil {
		return t.KeyPrefix + zreflect.ToString(key)
	}
	return t.KeyPrefix + t.KeyConvert(key)
}

func (t *Transformer[K, V]) keysAsSlice(keys []K) []string {
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, t.keyAsString(key))
	}
	return result
}

func (t *Transformer[K, V]) valueAsString(value V) (string, error) {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.String:
		return rv.String(), nil

	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return string(rv.Bytes()), nil
		}
	}

	return xcodec.EncodeToString(t.ValueCodec, value)
}

func (t *Transformer[K, V]) decodeValue(str string) (V, error) {
	var value V
	rv := reflect.ValueOf(&value).Elem()

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(str)
		return value, nil

	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			rv.SetBytes([]byte(str))
			return value, nil
		}
	}

	err := xcodec.DecodeFromString(t.ValueCodec, str, &value)
	return value, err
}

func (t *Transformer[K, V]) Has(ctx context.Context, key K) (bool, error) {
	return t.Cache.Has(ctx, t.keyAsString(key))
}

func (t *Transformer[K, V]) TTL(ctx context.Context, key K) (time.Duration, error) {
	return t.Cache.TTL(ctx, t.keyAsString(key))
}

func (t *Transformer[K, V]) Expire(ctx context.Context, key K, life time.Duration) error {
	return t.Cache.Expire(ctx, t.keyAsString(key), life)
}

func (t *Transformer[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	str, err := t.Cache.Get(ctx, t.keyAsString(key))
	if err != nil {
		return value, err
	}
	return t.decodeValue(str)
}

func (t *Transformer[K, V]) MGet(ctx context.Context, keys ...K) (result map[K]V, err error) {
	if mc, ok := t.Cache.(MGetter[string, string]); ok {
		strKeys := t.keysAsSlice(keys)
		rt, err1 := mc.MGet(ctx, strKeys...)
		if len(rt) == 0 {
			return result, err1
		}
		var errs []error
		if err1 != nil {
			errs = append(errs, err1)
		}

		mp := make(map[string]K, len(keys))
		for idx, strKey := range strKeys {
			mp[strKey] = keys[idx]
		}

		result = make(map[K]V, len(keys))
		for strKey, strVal := range rt {
			value, err2 := t.decodeValue(strVal)
			if err2 == nil {
				result[mp[strKey]] = value
			} else {
				errs = append(errs, err2)
			}
		}
		return result, errors.Join(errs...)
	}
	var errs []error
	result = make(map[K]V, len(keys))
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

func (t *Transformer[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	str, err := t.valueAsString(value)
	if err != nil {
		return err
	}
	return t.Cache.Set(ctx, t.keyAsString(key), str, ttl)
}

func (t *Transformer[K, V]) MSet(ctx context.Context, values map[K]V, ttl time.Duration) error {
	if mc, ok := t.Cache.(MSetter[string, string]); ok {
		var errs []error
		kv := make(map[string]string, len(values))
		for key, value := range values {
			strValue, err := t.valueAsString(value)
			if err == nil {
				kv[t.keyAsString(key)] = strValue
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

func (t *Transformer[K, V]) Delete(ctx context.Context, keys ...K) error {
	return t.Cache.Delete(ctx, t.keysAsSlice(keys)...)
}

var _ HasStats = (*Transformer[string, any])(nil)

func (t *Transformer[K, V]) Stats() Stats {
	if hs, ok := t.Cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{}
}

var _ HasAllStats = (*Transformer[string, any])(nil)

func (t *Transformer[K, V]) AllStats() map[string]Stats {
	if hs, ok := t.Cache.(HasAllStats); ok {
		return hs.AllStats()
	}
	return nil
}

func (t *Transformer[K, V]) Unwrap() any {
	return t.Cache
}
