package xkv

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

func checkSet(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	t.Run("set1", func(t *testing.T) {
		se := kvs.Set("set1")
		num, err := se.SAdd(ctx, "m1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = se.SAdd(ctx, "m1")
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		num, err = se.SAdd(ctx, "m1", "m2")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = se.SCard(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		gots, err := se.SMembers(ctx)
		xt.NoError(t, err)
		xt.Equal(t, gots, []string{"m1", "m2"})

		var values []string
		err = se.SRange(ctx, func(member string) bool {
			values = append(values, member)
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, values, []string{"m1", "m2"})

		err = se.SRem(ctx, "m2")
		xt.NoError(t, err)

		gots, err = se.SMembers(ctx)
		xt.NoError(t, err)
		xt.Equal(t, gots, []string{"m1"})
	})

	t.Run("set2", func(t *testing.T) {
		se := kvs.Set("t2-set2")

		ok, err := se.SIsMember(ctx, "m1")
		xt.NoError(t, err)
		xt.False(t, ok)

		num, err := se.SAdd(ctx, "m1", "m2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		for _, m := range []string{"m1", "m2"} {
			ok, err = se.SIsMember(ctx, m)
			xt.NoError(t, err)
			xt.True(t, ok)
		}

		oks, err := se.SMIsMember(ctx, []string{"m1", "m2", "m3-not-found"})
		xt.NoError(t, err)
		xt.Equal(t, oks, []bool{true, true, false})
	})

	t.Run("set3-pop", func(t *testing.T) {
		se := kvs.Set("t2-set3")
		members := []string{"m1", "m2", "m3", "m4"}
		num, err := se.SAdd(ctx, members...)
		xt.NoError(t, err)
		xt.Equal(t, num, 4)

		one, found, err := se.SPop(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.SliceContains(t, members, one)

		num, err = se.SCard(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 3)

		many, err := se.SPopN(ctx, 2)
		xt.NoError(t, err)
		xt.SliceContains(t, members, many...)

		num, err = se.SCard(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 1)
	})

	t.Run("set4-rand", func(t *testing.T) {
		se := kvs.Set("t2-set4")
		members := []string{"m1", "m2", "m3", "m4"}
		num, err := se.SAdd(ctx, members...)
		xt.NoError(t, err)
		xt.Equal(t, num, 4)

		one, found, err := se.SRandMember(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.SliceContains(t, members, one)

		num, err = se.SCard(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 4)

		many, err := se.SRandMemberN(ctx, 2)
		xt.NoError(t, err)
		xt.SliceContains(t, members, many...)

		num, err = se.SCard(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 4)
	})
}
