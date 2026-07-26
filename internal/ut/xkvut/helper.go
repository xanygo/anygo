package xkvut

import (
	"context"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
	"sync"
	"time"
)

var flags sync.Map

func ClearFlags() {
	flags.Range(func(k, v interface{}) bool {
		flags.Delete(k)
		return true
	})
}

func SetFlag(flag string) {
	flags.Store(flag, true)
}

func hasFlag(flag string) bool {
	_, ok := flags.Load(flag)
	return ok
}

func TestStringStorage1(t xt.TB, ff xkv.StringStorage) {
	t.Run("string", func(t xt.TB) {
		const key = "t1-hello"
		ss1 := ff.String(key)
		got1, found1, err1 := ss1.Get(context.Background())
		xt.NoError(t, err1)
		xt.False(t, found1)
		xt.Equal(t, got1, "")
		xt.NoError(t, ss1.Set(context.Background(), "world"))
		got2, found2, err2 := ss1.Get(context.Background())
		xt.True(t, found2)
		xt.NoError(t, err2)
		xt.Equal(t, got2, "world")

		got3, err3 := ff.Has(context.Background(), key)
		xt.NoError(t, err3)
		xt.True(t, got3)

		xt.NoError(t, ff.Delete(context.Background(), key))
	})

	t.Run("list", func(t xt.TB) {
		const key = "t1-list1"
		list := ff.List(key)
		_, err1 := list.RPush(context.Background(), "1")
		xt.NoError(t, err1)

		_, err2 := list.RPush(context.Background(), "2")
		xt.NoError(t, err2)
		var values []string
		err3 := list.RRange(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err3)
		xt.Equal(t, values, []string{"2", "1"})

		values = nil
		err4 := list.LRange(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err4)
		xt.Equal(t, values, []string{"1", "2"})

		values = nil
		err5 := list.Range(context.Background(), func(val string) bool {
			values = append(values, val)
			return true
		})
		xt.NoError(t, err5)
		xt.Len(t, values, 2)
	})

	t.Run("Hash", func(t xt.TB) {
		hh := ff.Hash("t1-hash1")
		xt.NoError(t, hh.HSet(context.Background(), "key1", "value1"))
		value1, found1, err1 := hh.HGet(context.Background(), "key1")
		xt.NoError(t, err1)
		xt.True(t, found1)
		xt.Equal(t, value1, "value1")

		value2, found2, err2 := hh.HGet(context.Background(), "key2")
		xt.NoError(t, err2)
		xt.False(t, found2)
		xt.Equal(t, value2, "")

		all, err4 := hh.HGetAll(context.Background())
		xt.NoError(t, err4)
		xt.Equal(t, all, map[string]string{"key1": "value1"})

		xt.NoError(t, hh.HDel(context.Background(), "key1"))
		value3, found3, err3 := hh.HGet(context.Background(), "key2")
		xt.NoError(t, err3)
		xt.False(t, found3)
		xt.Equal(t, value3, "")
	})

	t.Run("Set", func(t xt.TB) {
		set := ff.Set("t1-set1")
		_, err1 := set.SAdd(context.Background(), "v1")
		xt.NoError(t, err1)

		got1, err2 := set.SMembers(context.Background())
		xt.NoError(t, err2)
		xt.Equal(t, got1, []string{"v1"})

		_, err3 := set.SAdd(context.Background(), "v2")
		xt.NoError(t, err3)

		got2, err4 := set.SMembers(context.Background())
		xt.NoError(t, err4)
		xt.Equal(t, got2, []string{"v1", "v2"})

		xt.NoError(t, set.SRem(context.Background(), "v1"))
		got3, err3 := set.SMembers(context.Background())
		xt.NoError(t, err3)
		xt.Equal(t, got3, []string{"v2"})
	})

	t.Run("ZSet", func(t xt.TB) {
		zset := ff.ZSet("t1-zset1")
		xt.NoError(t, zset.ZAdd(context.Background(), 1, "m1"))
		got1, found1, err1 := zset.ZScore(context.Background(), "m1")
		xt.NoError(t, err1)
		xt.True(t, found1)
		xt.Equal(t, got1, 1)

		xt.NoError(t, zset.ZAdd(context.Background(), 2, "m2"))
		xt.NoError(t, zset.ZAdd(context.Background(), 1.5, "m3"))
		var members []string
		var scores []float64
		zset.ZRange(context.Background(), func(member string, score float64) bool {
			members = append(members, member)
			scores = append(scores, score)
			return true
		})
		xt.Equal(t, members, []string{"m1", "m3", "m2"})
		xt.Equal(t, scores, []float64{1, 1.5, 2})

		xt.NoError(t, zset.ZRem(context.Background(), "m2"))
		got2, found2, err2 := zset.ZScore(context.Background(), "m2")
		xt.NoError(t, err2)
		xt.False(t, found2)
		xt.Equal(t, got2, 0)
	})
}

func TestStringStorage2(t xt.TB, ff xkv.StringStorage) {
	checkString(t, ff)
	checkHash(t, ff)
	checkList(t, ff)
	checkSet(t, ff)
	checkZSet(t, ff)
}

func checkString(t xt.TB, xkv xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("k1", func(t xt.TB) {
		ks := xkv.String("t2-str-k1")

		t.Run("get1", func(t xt.TB) {
			value, found, err := ks.Get(ctx)
			xt.NoError(t, err)
			xt.False(t, found)
			xt.Empty(t, value)
		})

		t.Run("set1", func(t xt.TB) {
			err := ks.Set(ctx, "hello")
			xt.NoError(t, err)

			value, found, err := ks.Get(ctx)
			xt.NoError(t, err)
			xt.True(t, found)
			xt.Equal(t, value, "hello")
		})

		t.Run("incr1", func(t xt.TB) {
			num, err := ks.Incr(ctx)
			xt.Error(t, err)
			xt.Equal(t, num, 0)
		})
	})

	t.Run("k2", func(t xt.TB) {
		ks := xkv.String("t2-str-k2")
		checkGet := func(t xt.TB, want string) {
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
}

func checkList(t xt.TB, xkv xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("list1", func(t xt.TB) {
		l1 := xkv.List("t2-list1")
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

	t.Run("list2", func(t xt.TB) {
		li := xkv.List("t2-list2")
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

	t.Run("list3", func(t xt.TB) {
		li := xkv.List("t2-list3")
		num, err := li.RPush(ctx, "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)
	})

}

func checkHash(t xt.TB, xkv xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("hash1", func(t xt.TB) {
		ha := xkv.Hash("t2-hash1")
		value, found, err := ha.HGet(ctx, "f1")
		xt.NoError(t, err)
		xt.False(t, found)
		xt.Equal(t, value, "")

		err = ha.HSet(ctx, "f1", "v1")
		xt.NoError(t, err)

		value, found, err = ha.HGet(ctx, "f1")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "v1")

		err = ha.HSet(ctx, "f1", "v2")
		xt.NoError(t, err)

		value, found, err = ha.HGet(ctx, "f1")
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "v2")
	})

	t.Run("hash2", func(t xt.TB) {
		const key = "t2-hash2"
		ha := xkv.Hash(key)
		err := ha.HDel(ctx, "f1")
		xt.NoError(t, err)

		checkGet := func(t xt.TB, field string, want string) {
			t.Helper()
			value, found, err1 := ha.HGet(ctx, field)
			xt.NoError(t, err1)
			xt.Equal(t, value, want)
			xt.Equal(t, found, want != "")
		}

		vs := map[string]string{"f1": "v1", "f2": "v2"}
		err = ha.HMSet(ctx, vs)
		xt.NoError(t, err)

		checkGet(t, "f1", "v1")
		checkGet(t, "f2", "v2")
		checkGet(t, "f3", "")

		all := map[string]string{}
		err = ha.HRange(ctx, func(field string, value string) bool {
			all[field] = value
			return true
		})
		xt.NoError(t, err)
		xt.Equal(t, all, vs)

		got, err := ha.HGetAll(ctx)
		xt.NoError(t, err)
		xt.Equal(t, got, vs)

		err = ha.HDel(ctx, "f1", "f3")
		xt.NoError(t, err)
		checkGet(t, "f1", "")
		checkGet(t, "f2", "v2")
		checkGet(t, "f3", "")

		has, err := xkv.Has(ctx, key)
		xt.NoError(t, err)
		xt.True(t, has)
	})
}

func checkSet(t xt.TB, xkv xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("set1", func(t xt.TB) {
		se := xkv.Set("t2-set1")
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

		if hasFlag("SMembers-NotSorted") {
			xt.SliceSortEqual(t, gots, []string{"m1", "m2"})
		} else {
			xt.Equal(t, gots, []string{"m1", "m2"})
		}

		var values []string
		err = se.SRange(ctx, func(member string) bool {
			values = append(values, member)
			return true
		})
		xt.NoError(t, err)
		if hasFlag("SMembers-NotSorted") {
			xt.SliceSortEqual(t, values, []string{"m1", "m2"})
		} else {
			xt.Equal(t, values, []string{"m1", "m2"})
		}

		err = se.SRem(ctx, "m2")
		xt.NoError(t, err)

		gots, err = se.SMembers(ctx)
		xt.NoError(t, err)
		xt.Equal(t, gots, []string{"m1"})
	})
}

func checkZSet(t xt.TB, xkv xkv.StringStorage) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("zset1", func(t xt.TB) {
		zs := xkv.ZSet("t2-zset1")
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

	})

	t.Run("delete1", func(t xt.TB) {
		has, err := xkv.Has(ctx, "t2-zset1")
		xt.NoError(t, err)
		xt.True(t, has)

		err = xkv.Delete(ctx, "zset1")
		xt.NoError(t, err)

		has, err = xkv.Has(ctx, "zset1")
		xt.NoError(t, err)
		xt.False(t, has)
	})
}
