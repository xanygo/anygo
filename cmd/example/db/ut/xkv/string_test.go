package xkv

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

func checkString(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	t.Run("k1", func(t *testing.T) {
		ks := kvs.String("k1")

		t.Run("get1", func(t *testing.T) {
			value, found, err := ks.Get(ctx)
			xt.NoError(t, err)
			xt.False(t, found)
			xt.Empty(t, value)
		})

		t.Run("set1", func(t *testing.T) {
			err := ks.Set(ctx, "hello")
			xt.NoError(t, err)

			value, found, err := ks.Get(ctx)
			xt.NoError(t, err)
			xt.True(t, found)
			xt.Equal(t, value, "hello")
		})

		t.Run("incr1", func(t *testing.T) {
			num, err := ks.Incr(ctx)
			xt.Error(t, err)
			xt.Equal(t, num, 0)
		})

		t.Run("setnx", func(t *testing.T) {
			ok, err := ks.SetNX(ctx, "abc")
			xt.NoError(t, err)
			xt.False(t, ok)
		})
	})

	t.Run("setnx", func(t *testing.T) {
		ks2 := kvs.String("t2-setnx-1")

		ok, err := ks2.SetNX(ctx, "abc")
		xt.NoError(t, err)
		xt.True(t, ok)

		value, found, err := ks2.Get(ctx)
		xt.NoError(t, err)
		xt.Equal(t, value, "abc")
		xt.True(t, found)

		ok, err = ks2.SetNX(ctx, "hello")
		xt.NoError(t, err)
		xt.False(t, ok)

		value, found, err = ks2.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "abc")
	})

	t.Run("k2", func(t *testing.T) {
		ks := kvs.String("k2")
		checkGet := func(t *testing.T, want string) {
			t.Helper()
			value, found, err := ks.Get(ctx)
			xt.NoError(t, err)
			xt.True(t, found)
			xt.Equal(t, value, want)
		}
		for i := int64(0); i < 10; i++ {
			num, err := ks.Incr(ctx)
			xt.NoError(t, err)
			xt.Equal(t, num, 1+i)
		}

		checkGet(t, "10")

		for i := int64(0); i < 10; i++ {
			num, err := ks.Decr(ctx)
			xt.NoError(t, err)
			xt.Equal(t, num, 9-i)
		}
		checkGet(t, "0")
	})

	t.Run("k3", func(t *testing.T) {
		ks := kvs.String("t2-str-k3")

		num, err := ks.IncrBy(ctx, 2)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		value, found, err := ks.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "2")

		num, err = ks.IncrBy(ctx, 3)
		xt.NoError(t, err)
		xt.Equal(t, num, 5)

		value, found, err = ks.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "5")

		value, found, err = ks.GetDel(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "5")

		value, found, err = ks.GetDel(ctx) // 已被删除
		xt.NoError(t, err)
		xt.False(t, found)
		xt.Equal(t, value, "")
	})

	t.Run("k4-incrFloat", func(t *testing.T) {
		ks := kvs.String("t2-str-k4")
		num, err := ks.IncrByFloat(ctx, 1.2)
		xt.NoError(t, err)
		xt.Equal(t, num, 1.2)

		value, found, err := ks.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "1.2")

		num, err = ks.IncrByFloat(ctx, 2)
		xt.NoError(t, err)
		xt.Equal(t, num, 3.2)

		value, found, err = ks.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "3.2")
	})

	t.Run("k5-getset", func(t *testing.T) {
		ks := kvs.String("t2-str-k5")
		value, found, err := ks.GetSet(ctx, "hello")
		xt.NoError(t, err)
		xt.False(t, found)
		xt.Equal(t, value, "")

		value, found, err = ks.GetSet(ctx, "world")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "hello")

		value, found, err = ks.GetSet(ctx, "abc")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "world")
	})
}
