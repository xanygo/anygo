package xkv

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	_ "github.com/mattn/go-sqlite3"    // sqlite driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/xkvx"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xt"
)

var logWriter = &xt.TLogWriter{}

func init() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(logWriter)}).ToInterceptor())
}

func getDB(name string) *xdb.Client {
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		log.Fatalln(err)
	}

	return xdb.NewClient("sqlite3", "demo", db)
}

func TestSQLite(t *testing.T) {
	logWriter.Switch(t)

	db := getDB("ut.db")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %q", table))
		xt.NoError(t, err)
	}

	checkDB(t, db)
}

func checkDB(t *testing.T, db *xdb.Client) {
	kvs := &xkvx.DatabaseStore{
		DB: db,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xt.NoError(t, kvs.Migrate(ctx))

	t.Run("checkString", func(t *testing.T) {
		logWriter.Switch(t)
		checkString(t, kvs)
	})

	t.Run("checkList", func(t *testing.T) {
		logWriter.Switch(t)
		checkList(t, kvs)
	})

	t.Run("checkHash", func(t *testing.T) {
		logWriter.Switch(t)
		checkHash(t, kvs)
	})

	t.Run("checkSet", func(t *testing.T) {
		logWriter.Switch(t)
		checkSet(t, kvs)
	})

	t.Run("checkZSet", func(t *testing.T) {
		logWriter.Switch(t)
		checkZSet(t, kvs)
	})
}

var tables = []string{"xkv_meta", "xkv_string", "xkv_hash", "xkv_list", "xkv_set", "xkv_zset"}

func TestPostgres(t *testing.T) {
	logWriter.Switch(t)

	const dsn = `user=work password=123456 host=127.0.0.1 port=5432 database=demo sslmode=disable`
	db, err := sql.Open("pgx", dsn)
	xt.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		t.Skipf("PingContext failed: %s ,skipped", err.Error())
		return
	}

	client := xdb.NewClient("pgx", "demo", db)
	for _, table := range tables {
		_, err = xdb.Exec(ctx, client, fmt.Sprintf("DROP TABLE IF EXISTS %q", table))
		xt.NoError(t, err)
	}

	checkDB(t, client)
}

func TestMySQL(t *testing.T) {
	logWriter.Switch(t)
	checkMySQLBase(t, "mysql")
}

func TestMariaDB(t *testing.T) {
	logWriter.Switch(t)
	checkMySQLBase(t, "mariadb")
}

func checkMySQLBase(t *testing.T, dialect string) {
	const dsn = `work:123456@tcp(127.0.0.1)/demo`
	db, err := sql.Open("mysql", dsn)
	xt.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		t.Skipf("PingContext failed: %s ,skipped", err.Error())
		return
	}

	client := xdb.NewClient(dialect, "demo", db)
	for _, table := range tables {
		_, err = xdb.Exec(ctx, client, fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table))
		xt.NoError(t, err)
	}

	checkDB(t, client)
}

func checkString(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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

func checkList(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
}

func checkHash(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("hash1", func(t *testing.T) {
		ha := kvs.Hash("hash1")
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

	checkGetHa := func(t *testing.T, ha xkv.Hash[string], field string, want string) {
		t.Helper()
		value, found, err1 := ha.HGet(ctx, field)
		xt.NoError(t, err1)
		xt.Equal(t, value, want)
		xt.Equal(t, found, want != "")

		found2, err2 := ha.HExists(ctx, field)
		xt.NoError(t, err2)
		xt.Equal(t, found2, want != "")
	}

	t.Run("hash2", func(t *testing.T) {
		ha := kvs.Hash("hash2")
		err := ha.HDel(ctx, "f1")
		xt.NoError(t, err)

		checkGet := func(t *testing.T, field string, want string) {
			checkGetHa(t, ha, field, want)
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

		has, err := kvs.Has(ctx, "hash2")
		xt.NoError(t, err)
		xt.True(t, has)
	})

	t.Run("hash3-HIncrBy", func(t *testing.T) {
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

	t.Run("hash4-hlen", func(t *testing.T) {
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
}

func checkSet(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
}

func checkZSet(t *testing.T, kvs xkv.StringStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
	})

	t.Run("delete1", func(t *testing.T) {
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

		err = zs.ZAdd(ctx, 2, "f100") // 和 m2 相同的 score
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
}
