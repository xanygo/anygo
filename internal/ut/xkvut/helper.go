package xkvut

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

var flags sync.Map

func ClearFlags() {
	flags.Range(func(k, v any) bool {
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

func TestStringStorage1(t xt.TB, kvs xkv.StringStorage) {
	t.Run("String", func(t xt.TB) {
		const key = "t1-hello"
		ss1 := kvs.String(key)
		got1, found1, err1 := ss1.Get(context.Background())
		xt.NoError(t, err1)
		xt.False(t, found1)
		xt.Equal(t, got1, "")
		xt.NoError(t, ss1.Set(context.Background(), "world"))
		got2, found2, err2 := ss1.Get(context.Background())
		xt.True(t, found2)
		xt.NoError(t, err2)
		xt.Equal(t, got2, "world")

		got3, err3 := kvs.Has(context.Background(), key)
		xt.NoError(t, err3)
		xt.True(t, got3)

		xt.NoError(t, kvs.Delete(context.Background(), key))
	})

	t.Run("List", func(t xt.TB) {
		const key = "t1-list1"
		list := kvs.List(key)
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
		hh := kvs.Hash("t1-hash1")
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
		set := kvs.Set("t1-set1")
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
		zset := kvs.ZSet("t1-zset1")
		xt.NoError(t, zset.ZAdd(context.Background(), 1, "m1"))
		got1, found1, err1 := zset.ZScore(context.Background(), "m1")
		xt.NoError(t, err1)
		xt.True(t, found1)
		xt.Equal(t, got1, 1)

		xt.NoError(t, zset.ZAdd(context.Background(), 2, "m2"))
		xt.NoError(t, zset.ZAdd(context.Background(), 1.5, "m3"))
		var members []string
		var scores []float64
		err1 = zset.ZRange(context.Background(), func(member string, score float64) bool {
			members = append(members, member)
			scores = append(scores, score)
			return true
		})
		xt.NoError(t, err1)
		xt.SliceSortEqual(t, members, []string{"m1", "m3", "m2"})
		xt.SliceSortEqual(t, scores, []float64{1, 1.5, 2})

		xt.NoError(t, zset.ZRem(context.Background(), "m2"))
		got2, found2, err2 := zset.ZScore(context.Background(), "m2")
		xt.NoError(t, err2)
		xt.False(t, found2)
		xt.Equal(t, got2, 0)
	})
}

func TestStringStorage2(t xt.TB, kvs xkv.StringStorage) {
	t.Run("String", func(t xt.TB) {
		checkString(t, kvs)
	})

	t.Run("Hash", func(t xt.TB) {
		checkHash(t, kvs)
	})

	t.Run("List", func(t xt.TB) {
		checkList(t, kvs)
	})

	t.Run("Set", func(t xt.TB) {
		checkSet(t, kvs)
	})

	t.Run("ZSet", func(t xt.TB) {
		checkZSet(t, kvs)
	})
}

func checkString(t xt.TB, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("k1", func(t xt.TB) {
		ks := kvs.String("t2-str-k1")

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

		t.Run("setnx", func(t xt.TB) {
			ok, err := ks.SetNX(ctx, "abc")
			xt.NoError(t, err)
			xt.False(t, ok)

			ks2 := kvs.String("t2-str-k1-1")

			ok, err = ks2.SetNX(ctx, "abc")
			xt.NoError(t, err)
			xt.True(t, ok)

			value, found, err := ks2.Get(ctx)
			xt.NoError(t, err)
			xt.True(t, found)
			xt.Equal(t, value, "abc")

			ok, err = ks2.SetNX(ctx, "hello")
			xt.NoError(t, err)
			xt.False(t, ok)

			value, found, err = ks2.Get(ctx)
			xt.NoError(t, err)
			xt.True(t, found)
			xt.Equal(t, value, "abc")
		})
	})

	t.Run("k2", func(t xt.TB) {
		ks := kvs.String("t2-str-k2")
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

	t.Run("k3", func(t xt.TB) {
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

	t.Run("k4-incrFloat", func(t xt.TB) {
		ks := kvs.String("t2-str-k4")
		num, err := ks.IncrByFloat(ctx, 1.1)
		xt.NoError(t, err)
		xt.Equal(t, num, 1.1)

		value, found, err := ks.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "1.1")

		num, err = ks.IncrByFloat(ctx, 2)
		xt.NoError(t, err)
		xt.Equal(t, num, 3.1)

		value, found, err = ks.Get(ctx)
		xt.NoError(t, err)
		xt.True(t, found)
		xt.Equal(t, value, "3.1")
	})

	t.Run("k5-getset", func(t xt.TB) {
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

func checkList(t xt.TB, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("list1", func(t xt.TB) {
		l1 := kvs.List("t2-list1")
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
		li := kvs.List("t2-list2")
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
		li := kvs.List("t2-list3")
		num, err := li.RPush(ctx, "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)
	})

	t.Run("list4", func(t xt.TB) {
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

func checkHash(t xt.TB, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("hash1", func(t xt.TB) {
		ha := kvs.Hash("t2-hash1")
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

	checkGetHa := func(t xt.TB, ha xkv.Hash[string], field string, want string) {
		t.Helper()
		value, found, err1 := ha.HGet(ctx, field)
		xt.NoError(t, err1)
		xt.Equal(t, value, want)
		xt.Equal(t, found, want != "")

		found2, err2 := ha.HExists(ctx, field)
		xt.NoError(t, err2)
		xt.Equal(t, found2, want != "")
	}

	t.Run("hash2", func(t xt.TB) {
		const key = "t2-hash2"
		ha := kvs.Hash(key)
		err := ha.HDel(ctx, "f1")
		xt.NoError(t, err)

		vs := map[string]string{"f1": "v1", "f2": "v2"}
		err = ha.HMSet(ctx, vs)
		xt.NoError(t, err)

		checkGet := func(t xt.TB, field, want string) {
			checkGetHa(t, ha, field, want)
		}

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

		has, err := kvs.Has(ctx, key)
		xt.NoError(t, err)
		xt.True(t, has)
	})

	t.Run("hash3-HIncrBy", func(t xt.TB) {
		ha := kvs.Hash("t2-hash3")
		num, err := ha.HIncrBy(ctx, "f1", 1)
		xt.NoError(t, err)
		xt.Equal(t, num, 1)
		checkGetHa(t, ha, "f1", "1")

		num, err = ha.HIncrBy(ctx, "f1", 3)
		xt.NoError(t, err)
		xt.Equal(t, num, 4)
		checkGetHa(t, ha, "f1", "4")
	})

	t.Run("hash4-hlen", func(t xt.TB) {
		ha := kvs.Hash("t2-hash4")
		num, err := ha.HLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		for i := 0; i < 10; i++ {
			num, err = ha.HIncrBy(ctx, fmt.Sprintf("f-%d", i), int64(i))
			xt.NoError(t, err)
			xt.Equal(t, num, int64(i))

			num, err = ha.HLen(ctx)
			xt.NoError(t, err)
			xt.Equal(t, num, int64(i)+1)
		}
	})

	t.Run("hmget1", func(t xt.TB) {
		ha := kvs.Hash("t2-hmget1")
		result, err := ha.HMGet(ctx, "f1", "f2")
		xt.NoError(t, err)
		xt.Empty(t, result)
		kv1 := map[string]string{
			"f1": "v1",
			"f2": "v2",
			"f3": "v3",
		}
		err = ha.HMSet(ctx, kv1)
		xt.NoError(t, err)

		result, err = ha.HMGet(ctx, "f1", "f2")
		xt.NoError(t, err)
		xt.Equal(t, result, map[string]string{"f1": "v1", "f2": "v2"})
	})
}

func checkSet(t xt.TB, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("set1", func(t xt.TB) {
		se := kvs.Set("t2-set1")
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
		if hasFlag("SRange-NotSorted") {
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

	t.Run("set2", func(t xt.TB) {
		se := kvs.Set("t2-set2")

		ok, err := se.SIsMember(ctx, "m1")
		xt.NoError(t, err)
		xt.False(t, ok)

		num, err := se.SAdd(ctx, "m1", "m2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		ok, err = se.SIsMember(ctx, "m1")
		xt.NoError(t, err)
		xt.True(t, ok)

		oks, err := se.SMIsMember(ctx, []string{"m1", "m2", "m3-not-found"})
		xt.NoError(t, err)
		xt.Equal(t, oks, []bool{true, true, false})
	})

	t.Run("set3-pop", func(t xt.TB) {
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

	t.Run("set4-rand", func(t xt.TB) {
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

func checkZSet(t xt.TB, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("zset1", func(t xt.TB) {
		zs := kvs.ZSet("t2-zset1")
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
		has, err := kvs.Has(ctx, "t2-zset1")
		xt.NoError(t, err)
		xt.True(t, has)

		err = kvs.Delete(ctx, "zset1")
		xt.NoError(t, err)

		has, err = kvs.Has(ctx, "zset1")
		xt.NoError(t, err)
		xt.False(t, has)
	})

	t.Run("zincr", func(t xt.TB) {
		zs := kvs.ZSet("t2-zincr1")
		num, err := zs.ZIncrBy(ctx, 1.1, "m1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1.1)

		num, err = zs.ZIncrBy(ctx, 2.1, "m1")
		xt.NoError(t, err)
		xt.Equal(t, num, 3.2)
	})

	t.Run("zcount", func(t xt.TB) {
		zs := kvs.ZSet("t2-zcount1")

		for i := 0; i < 100; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}

		t.Run("inf", func(t xt.TB) {
			num, err := zs.ZCount(ctx, "-inf", "+inf")
			xt.NoError(t, err)
			xt.Equal(t, num, 100)
		})

		t.Run("1-5-1", func(t xt.TB) {
			num, err := zs.ZCount(ctx, "1", "5")
			xt.NoError(t, err)
			xt.Equal(t, num, 5)
		})

		t.Run("1-5-2", func(t xt.TB) {
			num, err := zs.ZCount(ctx, "(1", "(5")
			xt.NoError(t, err)
			xt.Equal(t, num, 3)
		})
	})

	t.Run("zrank", func(t xt.TB) {
		zs := kvs.ZSet("t2-rank1")
		for i := 0; i < 100; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}

		index, score, err := zs.ZRank(ctx, "m1")
		xt.NoError(t, err)
		xt.Equal(t, index, 1)
		xt.Equal(t, score, 1)

		index, score, err = zs.ZRank(ctx, "m99")
		xt.NoError(t, err)
		xt.Equal(t, index, 99)
		xt.Equal(t, score, 99)

		index, score, err = zs.ZRank(ctx, "10000")
		xt.NoError(t, err)
		xt.Equal(t, index, -1)
		xt.Equal(t, score, 0)

		err = zs.ZAdd(ctx, 2, "f100") // 和 m2 相同的 score,但是 f100 排在 m2 之前
		xt.NoError(t, err)

		index, score, err = zs.ZRank(ctx, "m2")
		xt.NoError(t, err)
		xt.Equal(t, index, 3)
		xt.Equal(t, score, 2)

		index, score, err = zs.ZRank(ctx, "f100")
		xt.NoError(t, err)
		xt.Equal(t, index, 2)
		xt.Equal(t, score, 2)
	})

	t.Run("zpopmax-min1", func(t xt.TB) {
		zs := kvs.ZSet("t2-zpopmax1")
		checkLen := func(t xt.TB, want int64) {
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

	t.Run("zrangebyscore", func(t xt.TB) {
		zs := kvs.ZSet("t2-zrangebyscore-1")
		for i := 0; i < 100; i++ {
			err := zs.ZAdd(ctx, float64(i), fmt.Sprintf("m%d", i))
			xt.NoError(t, err)
		}
		checkRange := func(t xt.TB, min, max string, want1 []string, wang2 []float64) {
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
}
