package xkv

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

func checkList(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	t.Run("list1", func(t *testing.T) {
		l1 := kvs.List("list1")
		var values []string
		err := l1.Range(ctx, func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err)
		xt.Empty(t, values)

		num, err := l1.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		num, err = l1.LRem(ctx, 0, "hello")
		xt.NoError(t, err)
		xt.Equal(t, num, 0)
	})

	t.Run("list2", func(t *testing.T) {
		li := kvs.List("list2")
		num, err := li.LPush(ctx, "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = li.LPush(ctx, "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		var values []string
		err = li.LRange(ctx, func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, values, []string{"v2", "v1"})

		values = nil
		err = li.RRange(ctx, func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, values, []string{"v1", "v2"})
	})

	t.Run("list3", func(t *testing.T) {
		li := kvs.List("list3")
		num, err := li.RPush(ctx, "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)
	})

	t.Run("list4", func(t *testing.T) {
		li := kvs.List("list4")
		result, err := li.LPopN(ctx, 3)
		xt.NoError(t, err)
		xt.Empty(t, result)

		num, err := li.RPush(ctx, "m1", "m2", "m3", "m4", "m5", "m6")
		xt.NoError(t, err)
		xt.Equal(t, num, 6)

		result, err = li.LPopN(ctx, 3)
		xt.NoError(t, err)
		xt.Equal(t, result, []string{"m1", "m2", "m3"})

		result, err = li.RPopN(ctx, 2)
		xt.NoError(t, err)
		xt.Equal(t, result, []string{"m6", "m5"})

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 1)
	})
}
