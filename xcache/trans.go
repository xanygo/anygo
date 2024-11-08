//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xcache

import (
	"context"
	"time"
	"unsafe"

	"github.com/xanygo/anygo/xcodec"
)

var _ Cache[string, any] = (*TransString[any])(nil)

type TransString[V any] struct {
	Cache Cache[string, string]
	Codec xcodec.Codec
}

func (t *TransString[V]) Get(ctx context.Context, key string) (value V, err error) {
	str, err := t.Cache.Get(ctx, key)
	if err != nil {
		return value, err
	}
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	err = t.Codec.Decode(bf, &value)
	return value, err
}

func (t *TransString[V]) Set(ctx context.Context, key string, value V, ttl time.Duration) error {
	bf, err := t.Codec.Encode(value)
	if err != nil {
		return err
	}
	str := unsafe.String(&bf[0], len(bf))
	return t.Cache.Set(ctx, key, str, ttl)
}

func (t *TransString[V]) Delete(ctx context.Context, keys ...string) error {
	return t.Cache.Delete(ctx, keys...)
}
