//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xcache

import (
	"context"
	"time"

	"github.com/xanygo/anygo/xcodec"
)

var _ Cache[string, any] = (*Transformer[any])(nil)
var _ HasStats = (*Transformer[any])(nil)

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

func (t *Transformer[V]) Set(ctx context.Context, key string, value V, ttl time.Duration) error {
	str, err := xcodec.EncodeToString(t.Codec, value)
	if err != nil {
		return err
	}
	return t.Cache.Set(ctx, key, str, ttl)
}

func (t *Transformer[V]) Delete(ctx context.Context, keys ...string) error {
	return t.Cache.Delete(ctx, keys...)
}

func (t *Transformer[V]) Stats() Stats {
	if hs, ok := t.Cache.(HasStats); ok {
		return hs.Stats()
	}
	return Stats{
		Keys: statsKeysNoStats,
	}
}
