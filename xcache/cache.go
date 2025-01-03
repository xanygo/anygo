//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"time"

	"github.com/xanygo/anygo/xerror"
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

func IsNotExists(err error) bool {
	return err != nil && xerror.IsNotFound(err)
}

type ValueError[V any] struct {
	Value V
	Err   error
}

const cacheFileExt = ".cache"
