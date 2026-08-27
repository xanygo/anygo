package modernc_sqlte

import (
	"database/sql"
	"os"
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"
	_ "modernc.org/sqlite" // sqlite driver

	"cmd/example/db/ut/model"
)

func TestSQLite(t *testing.T) {
	name := "./modernc_sqlte3.db"
	_ = os.Remove(name)
	db, err := sql.Open("sqlite", name)
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlite3", "demo", db)
	model.DoCheck(t, client)
}
