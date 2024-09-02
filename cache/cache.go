//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package cache

import (
	"context"
	"errors"
	"time"
)

type (
	Cache[K comparable, V any] interface {
		Getter[K, V]
		Setter[K, V]
		Deleter[K]
	}

	Getter[K comparable, V any] interface {
		Get(ctx context.Context, key K) (value V, err error)
	}

	Setter[K comparable, V any] interface {
		Set(ctx context.Context, key K, value V, ttl time.Duration) error
	}

	Deleter[K comparable] interface {
		Delete(ctx context.Context, keys ...K) error
	}
)

// ErrNil 缓存数据不存在
var ErrNil = errors.New("cache not exists")
