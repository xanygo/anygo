package xkv

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/xkvx"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xt"

	"cmd/example/db/internal"
)

var logWriter = &xt.TLogWriter{}

func init() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(logWriter)}).ToInterceptor())
	internal.Init()
}

var tables = []string{"xkv_meta", "xkv_string", "xkv_hash", "xkv_list", "xkv_set", "xkv_zset"}

func getSQLiteDB(name string) *xdb.Client {
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		log.Fatalln(err)
	}

	return xdb.NewClient("sqlite3", "demo", db)
}

func TestSQLite(t *testing.T) {
	logWriter.Switch(t)

	db := getSQLiteDB("ut.db")
	checkDB(t, db)
}

func TestMSSQL(t *testing.T) {
	logWriter.Switch(t)

	db, err := internal.NewMSSQL()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlserver", "demo", db)

	checkDB(t, client)
}

func checkDB(t *testing.T, db *xdb.Client) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil && internal.IsConnectFailedErr(err) {
		t.Skipf("Ping failed: %v", err)
		return
	}
	xt.NoError(t, err)

	kvs := &xkvx.Database{
		DB: db,
	}

	t.Run("drop table", func(s *testing.T) {
		sc := xdb.MustNewSchemaAPI(db)
		for _, table := range tables {
			err := sc.DropTableIfExists(ctx, table)
			xt.NoError(t, err)
		}
	})

	xt.NoError(t, kvs.Migrate(ctx))

	t.Run("schema", func(t *testing.T) {
		sc := xdb.MustNewSchemaAPI(db)
		t.Run("CurrentDatabase", func(t *testing.T) {
			name, err := sc.CurrentDatabase(ctx)
			xt.NoError(t, err)
			xt.NotEmpty(t, name)
		})
		t.Run("Tables", func(t *testing.T) {
			ts, err := sc.Tables(ctx)
			xt.NoError(t, err)
			xt.NotEmpty(t, ts)
			xt.SliceContains(t, ts, "xkv_meta")
		})
		t.Run("TableColumns", func(t *testing.T) {
			for _, table := range tables {
				cs, err := sc.TableColumns(ctx, table)
				xt.NoError(t, err)
				xt.NotEmpty(t, cs)
			}
		})
	})

	checkAll(t, kvs)
}

func checkAll(t *testing.T, kvs xkv.StringStorage) {
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

func TestPostgres(t *testing.T) {
	logWriter.Switch(t)

	db, err := internal.NewPostgres()
	xt.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		t.Skipf("PingContext failed: %s ,skipped", err.Error())
		return
	}

	client := xdb.NewClient("pgx", "demo", db)
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

	if err = db.PingContext(ctx); err != nil {
		t.Skipf("PingContext failed: %s ,skipped", err.Error())
		return
	}

	client := xdb.NewClient(dialect, "demo", db)

	checkDB(t, client)
}

func TestRedis(t *testing.T) {
	client := internal.NewRedis()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	err := client.Ping(ctx)
	if err != nil {
		t.Skip(err.Error())
		return
	}
	xt.NoError(t, client.Select(ctx, 10))
	xt.NoError(t, client.FlushB(ctx, true))

	xs := &xkvx.Redis{
		Client: client,
	}
	checkAll(t, xs)
}
