package xcache

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/store/xcache/xcachex"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xt"

	"cmd/example/db/internal"
)

var logWriter = &xt.TLogWriter{}

func init() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(logWriter)}).ToInterceptor())
	internal.Init()
}

func getSQLiteDB(name string) *xdb.Client {
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		log.Fatalln(err)
	}

	return xdb.NewClient("sqlite3", "cache", db)
}

var tables = []string{"xcache"}

func TestSQLite(t *testing.T) {
	logWriter.Switch(t)

	db := getSQLiteDB("xcache_ut.db")

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %q", table))
		xt.NoError(t, err)
	}
	checkDB(t, db)
}

func TestPostgres(t *testing.T) {
	logWriter.Switch(t)

	db, err := internal.NewPostgres()
	xt.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil && internal.IsConnectFailedErr(err) {
		t.Skipf("PingContext failed: %s ,skipped", err.Error())
		return
	}
	xt.NoError(t, err)

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
	db, err := internal.NewMySQL()
	xt.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil && internal.IsConnectFailedErr(err) {
		t.Skipf("PingContext failed: %s ,skipped", err.Error())
		return
	}
	xt.NoError(t, err)

	client := xdb.NewClient(dialect, "demo", db)
	for _, table := range tables {
		_, err = xdb.Exec(ctx, client, fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table))
		xt.NoError(t, err)
	}

	checkDB(t, client)
}

func checkDB(t *testing.T, db *xdb.Client) {
	cc := &xcachex.Database{
		DB: db,
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	xt.NoError(t, cc.Migrate(ctx))
	_, err := cc.ClearExpired(ctx, 100, 10)
	xt.NoError(t, err)

	checkAll(t, cc)
}

func checkAll(t *testing.T, cache xcache.StringCache) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	t.Run("single", func(t *testing.T) {
		got, err := cache.Get(ctx, "k1")
		xt.ErrorIs(t, err, xerror.NotFound)
		xt.Empty(t, got)

		has, err := cache.Has(ctx, "k1")
		xt.NoError(t, err)
		xt.False(t, has)

		err = cache.Set(ctx, "k1", "v1", time.Minute)
		xt.NoError(t, err)

		got, err = cache.Get(ctx, "k1")
		xt.NoError(t, err)
		xt.Equal(t, got, "v1")

		has, err = cache.Has(ctx, "k1")
		xt.NoError(t, err)
		xt.True(t, has)

		err = cache.Delete(ctx, "k1")
		xt.NoError(t, err)

		has, err = cache.Has(ctx, "k1")
		xt.NoError(t, err)
		xt.False(t, has)
	})

	t.Run("multi", func(t *testing.T) {
		mc, ok := cache.(xcache.StringMCache)
		if !ok {
			t.Skip("not mcache")
			return
		}
		kv1 := map[string]string{
			"k20": "v1",
			"k21": "v2",
			"k22": "v3",
		}
		err := mc.MSet(ctx, kv1, time.Minute)
		xt.NoError(t, err)

		got1, err := mc.MGet(ctx, "k20", "k21", "k22")
		xt.NoError(t, err)
		xt.Equal(t, got1, kv1)
	})
}
