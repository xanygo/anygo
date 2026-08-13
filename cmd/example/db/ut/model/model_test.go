package model

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xt"

	"cmd/example/db/internal"
)

func init() {
	internal.Init()
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(os.Stderr)}).ToInterceptor())
}

func TestSQLite(t *testing.T) {
	name := "./foo.db"
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlite3", "demo", db)
	doCheck(t, client)
}

func doCheck(t *testing.T, client *xdb.Client) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	err := client.PingContext(ctx)
	if err != nil {
		t.Skipf("ping db failed: %v", err)
		return
	}

	t.Run("withUser", func(t *testing.T) {
		withUser(ctx, t, client)
	})

	t.Run("withMPK", func(t *testing.T) {
		withMPK(ctx, t, client)
	})
}

func TestPGX(t *testing.T) {
	db, err := internal.NewPostgres()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("pgx", "demo", db)
	doCheck(t, client)
}

func TestMySQL(t *testing.T) {
	db, err := internal.NewMySQL()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("mysql", "demo", db)
	doCheck(t, client)
}

func TestMSSQL(t *testing.T) {
	db, err := internal.NewMSSQL()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlserver", "demo", db)
	doCheck(t, client)
}
