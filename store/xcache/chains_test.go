//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-28

package xcache_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/store/xcache"
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
	for i := 0; i < 10; i++ {
		value1, err1 := c.Get(ctx, "k1")
		fst.Error(t, err1)
		fst.True(t, xcache.IsNotExists(err1))
		fst.Equal(t, "", value1)
	}
	err2 := l2.Cache.Set(ctx, "k1", "v1", time.Minute)
	fst.NoError(t, err2)

	synctest.Test(t, func(t *testing.T) {
		value3, err3 := c.Get(ctx, "k1")
		fst.NoError(t, err3)
		fst.Equal(t, "v1", value3)

		synctest.Wait()

		value4, err4 := l1.Cache.Get(ctx, "k1")
		fst.NoError(t, err4)
		fst.Equal(t, "v1", value4)
	})

	err5 := c.Delete(ctx, "k1")
	fst.NoError(t, err5)

	value6, err6 := c.Get(ctx, "k1")
	fst.Error(t, err6)
	fst.True(t, xcache.IsNotExists(err6))
	fst.Equal(t, "", value6)
}
