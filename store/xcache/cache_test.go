//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xerror"
)

func testCache(t *testing.T, c Cache[string, int]) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	checkNotExists := func(t *testing.T) {
		got1, err1 := c.Get(ctx, "k1")
		fst.Equal(t, 0, got1)
		fst.ErrorIs(t, err1, xerror.NotFound)
	}
	checkNotExists(t)

	fst.NoError(t, c.Set(ctx, "k1", 1, 10*time.Second))
	got2, err2 := c.Get(ctx, "k1")
	fst.NoError(t, err2)
	fst.Equal(t, 1, got2)

	fst.NoError(t, c.Delete(ctx, "k1"))
	checkNotExists(t)
}
