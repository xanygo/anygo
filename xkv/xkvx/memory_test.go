//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-24

package xkvx_test

import (
	"fmt"
	"testing"

	"github.com/xanygo/anygo/xkv/xkvx"
	"github.com/xanygo/anygo/xt"
)

func TestMemory(t *testing.T) {
	ff := &xkvx.Memory{}
	testStringStorage(t, ff)
}

func BenchmarkMemory(b *testing.B) {
	ff := &xkvx.Memory{}
	benchStorage(b, ff)
}

func TestMemoryRange(t *testing.T) {
	mem := xkvx.NewMemory()

	t.Run("set", func(t *testing.T) {
		ls := mem.Set("set")
		for i := 0; i < 100; i++ {
			_, err := ls.SAdd(t.Context(), fmt.Sprintf("f-%d", i))
			xt.NoError(t, err)
		}
		err := ls.SRange(t.Context(), func(member string) bool {
			err := ls.SRem(t.Context(), member)
			xt.NoError(t, err)
			return true
		})
		xt.NoError(t, err)
		num, err := ls.SCard(t.Context())
		xt.NoError(t, err)
		xt.Equal(t, num, 0)
	})

	t.Run("hash", func(t *testing.T) {
		ls := mem.Hash("hash")
		for i := 0; i < 100; i++ {
			err := ls.HSet(t.Context(), fmt.Sprintf("f-%d", i), "hello")
			xt.NoError(t, err)
		}
		err := ls.HRange(t.Context(), func(field, value string) bool {
			err := ls.HDel(t.Context(), field)
			xt.NoError(t, err)
			return true
		})
		xt.NoError(t, err)
		num, err := ls.HLen(t.Context())
		xt.NoError(t, err)
		xt.Equal(t, num, 0)
	})

	t.Run("list", func(t *testing.T) {
		ls := mem.List("list")
		for i := 0; i < 100; i++ {
			_, err := ls.LPush(t.Context(), "hello")
			xt.NoError(t, err)
		}
		err := ls.LRange(t.Context(), func(val string) bool {
			_, err := ls.LRem(t.Context(), 1, val)
			xt.NoError(t, err)
			return true
		})
		xt.NoError(t, err)
		num, err := ls.LLen(t.Context())
		xt.NoError(t, err)
		xt.Equal(t, num, 0)
	})

	t.Run("zset", func(t *testing.T) {
		zs := mem.ZSet("zset")
		for i := 0; i < 100; i++ {
			err := zs.ZAdd(t.Context(), float64(i), fmt.Sprintf("a-%d", i))
			xt.NoError(t, err)
		}
		err := zs.ZRange(t.Context(), func(member string, score float64) bool {
			err := zs.ZRem(t.Context(), member)
			xt.NoError(t, err)
			return true
		})
		xt.NoError(t, err)
		num, err := zs.ZLen(t.Context())
		xt.NoError(t, err)
		xt.Equal(t, num, 0)
	})

}
