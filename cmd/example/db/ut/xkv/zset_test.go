package xkv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

func checkZSet(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	t.Run("zset1", func(t *testing.T) {
		zs := kvs.ZSet("zset1")
		err := zs.ZAdd(ctx, 1, "m1")
		xt.NoError(t, err)

		err = zs.ZAdd(ctx, 2, "m1")
		xt.NoError(t, err)

		err = zs.ZAdd(ctx, 3, "m2")
		xt.NoError(t, err)

		score, found, err := zs.ZScore(ctx, "m1")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, score, 2)

		err = zs.ZRem(ctx, "m1")
		xt.NoError(t, err)

		score, found, err = zs.ZScore(ctx, "m1")
		xt.NoError(t, err)
		xt.False(t, found)
		xt.Equal(t, score, 0)

		zl, err := zs.ZLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, zl, 1)

		has, err := kvs.Has(ctx, "zset1")
		xt.NoError(t, err)
		xt.True(t, has)

		err = kvs.Delete(ctx, "zset1")
		xt.NoError(t, err)

		has, err = kvs.Has(ctx, "zset1")
		xt.NoError(t, err)
		xt.False(t, has)
	})

	t.Run("zset2", func(t *testing.T) {
		zs := kvs.ZSet("zset2")
		for i := 0; i < 10; i++ {
			member := fmt.Sprintf("m%d", i)
			err := zs.ZAdd(ctx, float64(i)+10.1, member)
			xt.NoError(t, err)
		}
	})

	t.Run("zincr", func(t *testing.T) {
		zs := kvs.ZSet("t1-zincr1")
		num, err := zs.ZIncrBy(ctx, 1.1, "m1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1.1)

		num, err = zs.ZIncrBy(ctx, 2.1, "m1")
		xt.NoError(t, err)
		xt.Equal(t, num, 3.2)
	})

	t.Run("zcount", func(t *testing.T) {
		zs := kvs.ZSet("t1-zcount1")

		for i := 0; i < 100; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}

		num, err := zs.ZCount(ctx, "-inf", "+inf")
		xt.NoError(t, err)
		xt.Equal(t, num, 100)

		num, err = zs.ZCount(ctx, "1", "5")
		xt.NoError(t, err)
		xt.Equal(t, num, 5)

		num, err = zs.ZCount(ctx, "(1", "(5")
		xt.NoError(t, err)
		xt.Equal(t, num, 3)
	})

	t.Run("zrank", func(t *testing.T) {
		zs := kvs.ZSet("t2-rank1")
		for i := 0; i < 100; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}

		index, score, err := zs.ZRank(ctx, "m1")
		xt.NoError(t, err)
		xt.Equal(t, score, 1)
		xt.Equal(t, index, 1)

		index, score, err = zs.ZRank(ctx, "m99")
		xt.NoError(t, err)
		xt.Equal(t, score, 99)
		xt.Equal(t, index, 99)

		index, score, err = zs.ZRank(ctx, "10000")
		xt.NoError(t, err)
		xt.Equal(t, score, 0)
		xt.Equal(t, index, -1)

		err = zs.ZAdd(ctx, 2, "f100") // 和 m2 相同的 score
		xt.NoError(t, err)

		indexM2, score, err := zs.ZRank(ctx, "m2")
		xt.NoError(t, err)
		xt.Equal(t, score, 2)

		// member 值 hash 后 f100，排在 m2 之后
		// redis 存储引擎的值是2，数据库存储的值是3
		xt.True(t, indexM2 == 2 || indexM2 == 3)

		indexF100, score, err := zs.ZRank(ctx, "f100")
		xt.NoError(t, err)
		xt.Equal(t, score, 2)

		// member 值 hash 后 f100，排在 m2 之后
		// redis 存储引擎的值是3，数据库存储的值是3
		xt.True(t, indexF100 == 2 || indexF100 == 3)

		xt.NotEqual(t, indexM2, indexF100)
	})

	t.Run("zpopmax-min1", func(t *testing.T) {
		zs := kvs.ZSet("t2-zpopmax1")
		checkLen := func(t *testing.T, want int64) {
			t.Helper()
			num, err1 := zs.ZLen(ctx)
			xt.NoError(t, err1)
			xt.Equal(t, num, want)
		}
		members, scores, err := zs.ZPopMax(ctx, 3)
		xt.NoError(t, err)
		xt.Empty(t, members)
		xt.Empty(t, scores)

		err = zs.ZAdd(ctx, 1, "m1")
		xt.NoError(t, err)

		checkLen(t, 1)

		members, scores, err = zs.ZPopMax(ctx, 3)
		xt.NoError(t, err)
		xt.Equal(t, members, []string{"m1"})
		xt.Equal(t, scores, []float64{1})

		checkLen(t, 0)

		for i := 0; i < 20; i++ {
			err = zs.ZAdd(ctx, float64(i+1), fmt.Sprintf("m%d", i+1))
			xt.NoError(t, err)
		}
		checkLen(t, 20)

		members, scores, err = zs.ZPopMax(ctx, 3)
		xt.NoError(t, err)
		xt.Equal(t, members, []string{"m20", "m19", "m18"})
		xt.Equal(t, scores, []float64{20, 19, 18})

		checkLen(t, 17)

		members, scores, err = zs.ZPopMax(ctx, 3)
		xt.NoError(t, err)
		xt.Equal(t, members, []string{"m17", "m16", "m15"})
		xt.Equal(t, scores, []float64{17, 16, 15})

		checkLen(t, 14)

		members, scores, err = zs.ZPopMin(ctx, 4)
		xt.NoError(t, err)
		xt.Equal(t, members, []string{"m1", "m2", "m3", "m4"})
		xt.Equal(t, scores, []float64{1, 2, 3, 4})

		checkLen(t, 10)
	})

	t.Run("zrangebyscore", func(t *testing.T) {
		zs := kvs.ZSet("t2-zrangebyscore-1")
		for i := 0; i < 100; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}
		checkRange := func(t *testing.T, min, max string, want1 []string, wang2 []float64) {
			var members []string
			var scores []float64
			err := zs.ZRangeByScore(ctx, min, max, func(member string, score float64) bool {
				members = append(members, member)
				scores = append(scores, score)
				return true
			})
			xt.NoError(t, err)
			xt.SliceSortEqual(t, members, want1)
			xt.SliceSortEqual(t, scores, wang2)
		}

		checkRange(t, "1", "2", []string{"m1", "m2"}, []float64{1, 2})
	})

	t.Run("zrem-rangebyscore", func(t *testing.T) {
		zs := kvs.ZSet("t2-zrem-rangebyscore-1")
		for i := 0; i < 10; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}
		num, err := zs.ZLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 10)

		num, err = zs.ZRemRangeByScore(ctx, "-inf", "3")
		xt.NoError(t, err)
		xt.Equal(t, num, 4)

		num, err = zs.ZLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 6)
	})
}
