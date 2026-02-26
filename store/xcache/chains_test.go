//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-28

package xcache_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xt"
)

func TestNewChains(t *testing.T) {
	l1 := &xcache.Chain[string, string]{
		Cache: xcache.NewLRU[string, string](10),
		DynamicTTL: func(ctx context.Context, key string, value string) time.Duration {
			return time.Minute
		},
	}
	l2 := &xcache.Chain[string, string]{
		Cache: xcache.NewLRU[string, string](10),
		DynamicTTL: func(ctx context.Context, key string, value string) time.Duration {
			return 2 * time.Minute
		},
	}
	c := xcache.NewChains(l1, l2)
	ctx := context.Background()
	for range 10 {
		value1, err1 := c.Get(ctx, "k1")
		xt.Error(t, err1)
		xt.True(t, xcache.IsNotExists(err1))
		xt.Equal(t, "", value1)
	}
	err2 := l2.Cache.Set(ctx, "k1", "v1", time.Minute)
	xt.NoError(t, err2)

	synctest.Test(t, func(t *testing.T) {
		value3, err3 := c.Get(ctx, "k1")
		xt.NoError(t, err3)
		xt.Equal(t, "v1", value3)

		synctest.Wait()

		value4, err4 := l1.Cache.Get(ctx, "k1")
		xt.NoError(t, err4)
		xt.Equal(t, "v1", value4)
	})

	err5 := c.Delete(ctx, "k1")
	xt.NoError(t, err5)

	value6, err6 := c.Get(ctx, "k1")
	xt.Error(t, err6)
	xt.True(t, xcache.IsNotExists(err6))
	xt.Equal(t, "", value6)
}
