//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-26

package xcache

import (
	"context"
	"errors"
	"time"

	"github.com/xanygo/anygo/xerror"
)

func NewChains[K comparable, V any](caches ...Cache[K, V]) Cache[K, V] {
	return &chains[K, V]{
		caches: caches,
	}
}

var _ StringCache = (*chains[string, string])(nil)

type chains[K comparable, V any] struct {
	caches []Cache[K, V]
}

func (c *chains[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	var errs []error
	for _, cache := range c.caches {
		value, err := cache.Get(ctx, key)
		if err == nil {
			return value, nil
		} else if err != nil && !xerror.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return v, errors.Join(errs...)
	}
	return v, xerror.NotFound
}

func (c *chains[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	var errs []error
	for _, cache := range c.caches {
		if err := cache.Set(ctx, key, value, ttl); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (c *chains[K, V]) Delete(ctx context.Context, keys ...K) error {
	var errs []error
	for _, cache := range c.caches {
		if err := cache.Delete(ctx, keys...); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
