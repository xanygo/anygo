package xkv

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv/xkvx"
	"github.com/xanygo/anygo/xt"
	"log"
	"os"
	"testing"
	"time"
)

func getDB(name string) *xdb.Client {
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		log.Fatalln(err)
	}

	return xdb.NewClient("sqlite3", "demo", db)
}

func TestString(t *testing.T) {
	db := getDB("string.db")
	xkv := xkvx.DatabaseStorage{
		DB: db,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, xkv.Migrate(ctx))

	t.Run("k1", func(t *testing.T) {
		ks := xkv.String("k1")

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
	})

	t.Run("k2", func(t *testing.T) {
		ks := xkv.String("k2")
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
}

func TestList(t *testing.T) {
	db := getDB("list.db")
	xkv := xkvx.DatabaseStorage{
		DB: db,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, xkv.Migrate(ctx))

	t.Run("list1", func(t *testing.T) {
		l1 := xkv.List("list1")
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
		li := xkv.List("list2")
		num, err := li.LPush(ctx, "v1")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		num, err = li.LPush(ctx, "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

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
		li := xkv.List("list3")
		num, err := li.RPush(ctx, "v1", "v2")
		xt.NoError(t, err)
		xt.Equal(t, num, 2)

		num, err = li.LLen(ctx)
		xt.NoError(t, err)
		xt.Equal(t, num, 2)
	})

}

func TestHash(t *testing.T) {
	db := getDB("hash.db")
	xkv := xkvx.DatabaseStorage{
		DB: db,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, xkv.Migrate(ctx))

	t.Run("hash1", func(t *testing.T) {
		ha := xkv.Hash("hash1")
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

	t.Run("hash2", func(t *testing.T) {
		ha := xkv.Hash("hash2")
		err := ha.HDel(ctx, "f1")
		xt.NoError(t, err)

		checkGet := func(t *testing.T, field string, want string) {
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

		has, err := xkv.Has(ctx, "hash2")
		xt.NoError(t, err)
		xt.True(t, has)
	})
}

func TestSet(t *testing.T) {
	db := getDB("set.db")
	xkv := xkvx.DatabaseStorage{
		DB: db,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, xkv.Migrate(ctx))

	t.Run("set1", func(t *testing.T) {
		se := xkv.Set("set1")
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
}

func TestZSet(t *testing.T) {
	db := getDB("zset.db")
	xkv := xkvx.DatabaseStorage{
		DB: db,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, xkv.Migrate(ctx))

	t.Run("zset1", func(t *testing.T) {
		zs := xkv.ZSet("zset1")
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

	t.Run("delete1", func(t *testing.T) {
		has, err := xkv.Has(ctx, "zset1")
		xt.NoError(t, err)
		xt.True(t, has)

		err = xkv.Delete(ctx, "zset1")
		xt.NoError(t, err)

		has, err = xkv.Has(ctx, "zset1")
		xt.NoError(t, err)
		xt.False(t, has)
	})
}
