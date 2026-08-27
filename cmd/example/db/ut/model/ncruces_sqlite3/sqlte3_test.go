package ncruces_sqlite3

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/ncruces/go-sqlite3/driver" // sqlite driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"

	"cmd/example/db/ut/model"
)

func TestSQLite(t *testing.T) {
	name := "./ncruces_sqlte3.db"
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlite3", "demo", db)
	model.DoCheck(t, client)
}
