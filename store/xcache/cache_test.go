//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func testCache(t *testing.T, c Cache[string, int]) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	checkNotExists := func(t *testing.T) {
		t.Helper()
		got1, err1 := c.Get(ctx, "k1")
		xt.Equal(t, 0, got1)
		xt.ErrorIs(t, err1, xerror.NotFound)
	}
	checkNotExists(t)

	xt.NoError(t, c.Set(ctx, "k1", 1, 10*time.Second))
	got2, err2 := c.Get(ctx, "k1")
	xt.NoError(t, err2)
	xt.Equal(t, 1, got2)

	xt.NoError(t, c.Delete(ctx, "k1"))
	checkNotExists(t)
}
