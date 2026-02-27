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
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	checkNotExists := func(t *testing.T) {
		t.Helper()
		got1, err1 := c.Get(ctx, "k1")
		xt.Equal(t, 0, got1)
		xt.ErrorIs(t, err1, xerror.NotFound)
		got2, err2 := c.Has(ctx, "k1")
		xt.False(t, got2)
		xt.NoError(t, err2)
	}
	t.Logf("checkNotExists-0")
	checkNotExists(t)

	t.Logf("check set k1")
	xt.NoError(t, c.Set(ctx, "k1", 1, 10*time.Second))

	t.Logf("check get k1")
	got2, err2 := c.Get(ctx, "k1")
	xt.NoError(t, err2)
	xt.Equal(t, 1, got2)

	t.Logf("check has k1")
	got3, err3 := c.Has(ctx, "k1")
	xt.NoError(t, err3)
	xt.True(t, got3)

	xt.NoError(t, c.Delete(ctx, "k1"))

	t.Logf("checkNotExists-1")
	checkNotExists(t)
}
