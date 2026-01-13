//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-13

package xcache_test

import (
	"net"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xt"
)

func TestIsNop(t *testing.T) {
	t.Run("is nop", func(t *testing.T) {
		c1 := &xcache.Nop[string, string]{}
		xt.True(t, xcache.IsNop(c1))

		c2 := xcache.NewLatencyObserver[string, string](c1, time.Hour, time.Minute)
		xt.True(t, xcache.IsNop(c2))

		c3 := &xcache.Transformer[net.IP]{
			Cache: c1,
		}
		xt.True(t, xcache.IsNop(c3))

		c4 := &xcache.Transformer[net.IP]{
			Cache: c2,
		}
		xt.True(t, xcache.IsNop(c4))
	})

	t.Run("not nop", func(t *testing.T) {
		c1 := &xcache.File[string, string]{}
		xt.False(t, xcache.IsNop(c1))

		c2 := xcache.NewLatencyObserver[string, string](c1, time.Hour, time.Minute)
		xt.False(t, xcache.IsNop(c2))

		c3 := &xcache.Transformer[net.IP]{
			Cache: c1,
		}
		xt.False(t, xcache.IsNop(c3))

		c4 := &xcache.Transformer[net.IP]{
			Cache: c2,
		}
		xt.False(t, xcache.IsNop(c4))
	})

	t.Run("chain", func(t *testing.T) {
		l1 := &xcache.Chain[string, string]{
			Cache: &xcache.Nop[string, string]{},
		}
		l2 := &xcache.Chain[string, string]{
			Cache: xcache.NewLRU[string, string](10),
		}
		chs := xcache.NewChains[string, string](l1, l2)
		xt.False(t, xcache.IsNop(chs))
	})
}
