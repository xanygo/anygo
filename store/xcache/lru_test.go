//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func TestLRU(t *testing.T) {
	c1 := NewLRU[string, int](10)
	testCache(t, c1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for i := 0; i < 11; i++ {
		xt.NoError(t, c1.Set(ctx, fmt.Sprintf("k_%d", i), i, 10*time.Second))
	}

	_, err1 := c1.Get(ctx, "k_0")
	xt.ErrorIs(t, err1, xerror.NotFound)

	got2, err2 := c1.Get(ctx, "k_1")
	xt.NoError(t, err2)
	xt.Equal(t, 1, got2)
}
