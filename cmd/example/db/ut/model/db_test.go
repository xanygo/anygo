package model

import (
	"os"
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xt"

	"cmd/example/db/internal"
)

func init() {
	internal.Init()
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(os.Stderr)}).ToInterceptor())
}

func TestPGX(t *testing.T) {
	db, err := internal.NewPostgres()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("pgx", "demo", db)
	DoCheck(t, client)
}

func TestMySQL(t *testing.T) {
	db, err := internal.NewMySQL()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("mysql", "demo", db)
	DoCheck(t, client)
}

func TestMSSQL(t *testing.T) {
	db, err := internal.NewMSSQL()
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlserver", "demo", db)
	DoCheck(t, client)
}
