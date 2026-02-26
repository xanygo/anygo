//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

func TestLRU(t *testing.T) {
	c1 := xcache.NewLRU[string, int](10)
	testCache(t, c1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var allKeys []string
	for i := range 11 {
		key := fmt.Sprintf("k_%d", i)
		xt.NoError(t, c1.Set(ctx, key, i, 10*time.Second))
		allKeys = append(allKeys, key)
	}

	var keys []string
	c1.RangeLocked(func(item *xcache.MemValue[string, int]) (remove bool, goon bool) {
		keys = append(keys, item.Key)
		return false, true
	})
	xt.Len(t, keys, 10)
	xt.Equal(t, xslice.TailN(allKeys, 10), keys)

	_, err1 := c1.Get(ctx, "k_0")
	xt.ErrorIs(t, err1, xerror.NotFound)

	got2, err2 := c1.Get(ctx, "k_1")
	xt.NoError(t, err2)
	xt.Equal(t, 1, got2)

	xt.NoError(t, c1.Set(ctx, fmt.Sprintf("k_%d", 100), 100, 10*time.Second))

	_, err3 := c1.Get(ctx, "k_2")
	xt.ErrorIs(t, err3, xerror.NotFound)

	xt.Equal(t, 10, c1.Count())
	c1.RangeLocked(func(item *xcache.MemValue[string, int]) (remove bool, goon bool) {
		return true, true
	})
	xt.Equal(t, 0, c1.Count())
}

func TestMemoryFIFO(t *testing.T) {
	c1 := xcache.NewMemoryFIFO[string, int](10)
	testCache(t, c1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var allKeys []string
	for i := range 11 {
		key := fmt.Sprintf("k_%d", i)
		xt.NoError(t, c1.Set(ctx, key, i, 10*time.Second))
		allKeys = append(allKeys, key)
	}

	var keys []string
	c1.RangeLocked(func(item *xcache.MemValue[string, int]) (remove bool, goon bool) {
		keys = append(keys, item.Key)
		return false, true
	})
	xt.Len(t, keys, 10)
	xt.Equal(t, xslice.TailN(allKeys, 10), keys)

	_, err1 := c1.Get(ctx, "k_0")
	xt.ErrorIs(t, err1, xerror.NotFound)

	got2, err2 := c1.Get(ctx, "k_1")
	xt.NoError(t, err2)
	xt.Equal(t, 1, got2)

	xt.NoError(t, c1.Set(ctx, fmt.Sprintf("k_%d", 100), 100, 10*time.Second))

	got3, err3 := c1.Get(ctx, "k_2")
	xt.NoError(t, err3)
	xt.Equal(t, 2, got3)

	xt.Equal(t, 10, c1.Count())
	c1.RangeLocked(func(item *xcache.MemValue[string, int]) (remove bool, goon bool) {
		return true, true
	})
	xt.Equal(t, 0, c1.Count())
}

func TestIsMemory(t *testing.T) {
	t.Run("is memory", func(t *testing.T) {
		c1 := xcache.NewLRU[string, string](1)
		xt.True(t, xcache.IsMemory(c1))

		c2 := xcache.NewLatencyObserver[string, string](c1, time.Hour, time.Minute)
		xt.True(t, xcache.IsMemory(c2))

		c3 := &xcache.Transformer[net.IP]{
			Cache: c1,
		}
		xt.True(t, xcache.IsMemory(c3))

		c4 := &xcache.Transformer[net.IP]{
			Cache: c2,
		}
		xt.True(t, xcache.IsMemory(c4))

		c5 := xcache.NewMemoryLIFO[string, string](1)
		xt.True(t, xcache.IsMemory(c5))
	})

	t.Run("nop not memory", func(t *testing.T) {
		c1 := &xcache.Nop[string, string]{}
		xt.False(t, xcache.IsMemory(c1))

		c2 := xcache.NewLatencyObserver[string, string](c1, time.Hour, time.Minute)
		xt.False(t, xcache.IsMemory(c2))

		c3 := &xcache.Transformer[net.IP]{
			Cache: c1,
		}
		xt.False(t, xcache.IsMemory(c3))

		c4 := &xcache.Transformer[net.IP]{
			Cache: c2,
		}
		xt.False(t, xcache.IsMemory(c4))

		c5 := &xcache.Transformer[net.IP]{}
		xt.False(t, xcache.IsMemory(c5))
	})

	t.Run("not memory", func(t *testing.T) {
		c1 := &xcache.File[string, string]{}
		xt.False(t, xcache.IsMemory(c1))

		c2 := xcache.NewLatencyObserver[string, string](c1, time.Hour, time.Minute)
		xt.False(t, xcache.IsMemory(c2))

		c3 := &xcache.Transformer[net.IP]{
			Cache: c1,
		}
		xt.False(t, xcache.IsMemory(c3))

		c4 := &xcache.Transformer[net.IP]{
			Cache: c2,
		}
		xt.False(t, xcache.IsMemory(c4))
	})

	t.Run("chain", func(t *testing.T) {
		l1 := &xcache.Chain[string, string]{
			Cache: &xcache.Nop[string, string]{},
		}
		l2 := &xcache.Chain[string, string]{
			Cache: xcache.NewLRU[string, string](10),
		}
		chs := xcache.NewChains[string, string](l1, l2)
		xt.True(t, xcache.IsMemory(chs))
	})
}
