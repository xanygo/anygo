//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func testCache(t *testing.T, c xcache.Cache[string, int]) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	checkNotExists := func(t *testing.T) {
		got1, err1 := c.Get(ctx, "k1")
		xt.Equal(t, got1, 0)
		xt.ErrorIs(t, err1, xerror.NotFound)
		got2, err2 := c.Has(ctx, "k1")
		xt.False(t, got2)
		xt.NoError(t, err2)
	}
	t.Logf("checkNotExists-0")
	checkNotExists(t)

	t.Logf("check set k1")
	xt.NoError(t, c.Set(ctx, "k1", 1, 10*time.Second))

	t.Run("ttl1", func(t *testing.T) {
		ttl1, err1 := c.TTL(ctx, "k1")
		xt.NoError(t, err1)
		xt.Greater(t, ttl1, 9*time.Second)
		xt.LessOrEqual(t, ttl1, 10*time.Second)
	})

	xt.NoError(t, c.Expire(ctx, "k1", time.Hour))

	t.Run("ttl2", func(t *testing.T) {
		ttl1, err1 := c.TTL(ctx, "k1")
		xt.NoError(t, err1)
		xt.Greater(t, ttl1, 59*time.Minute)
		xt.LessOrEqual(t, ttl1, time.Hour)
	})

	t.Run("ttl3", func(t *testing.T) {
		ttl1, err1 := c.TTL(ctx, "k-ttl-not-exists")
		xt.NoError(t, err1)
		xt.Equal(t, ttl1, 0)
	})

	t.Logf("check get k1")
	got2, err2 := c.Get(ctx, "k1")
	xt.NoError(t, err2)
	xt.Equal(t, got2, 1)

	t.Logf("check has k1")
	got3, err3 := c.Has(ctx, "k1")
	xt.NoError(t, err3)
	xt.True(t, got3)

	xt.NoError(t, c.Delete(ctx, "k1"))

	t.Run("key-not-exists", func(t *testing.T) {
		const key = "k-expire-not-exists"
		ttl1, err1 := c.TTL(ctx, key)
		xt.NoError(t, err1)
		xt.Equal(t, ttl1, 0)

		xt.Error(t, c.Expire(ctx, key, time.Hour))

		ttl1, err1 = c.TTL(ctx, key)
		xt.NoError(t, err1)
		xt.Equal(t, ttl1, 0)
	})

	t.Logf("checkNotExists-1")
	checkNotExists(t)
}
