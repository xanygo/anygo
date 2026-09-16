package xcookiejar_test

import (
	"database/sql"
	"net/http"
	"net/url"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/xanygo/anygo/xdb"
	"github.com/xanygo/anygo/xhttp/xcookiejar"
	"github.com/xanygo/anygo/xt"
)

func TestCookJar(t *testing.T) {
	name := "./jar_mattn_sqlte3.db"
	_ = os.Remove(name)
	db, err := sql.Open("sqlite3", name)
	xt.NoError(t, err)

	defer db.Close()
	client := xdb.NewClient("sqlite3", "demo", db)
	testDB(t, client)
}

func testDB(t *testing.T, client *xdb.Client) {
	const table = "xcookiejar"
	sc := xdb.MustNewSchemaAPI(client)
	err := sc.DropTableIfExists(t.Context(), table)
	xt.NoError(t, err)

	store := &xcookiejar.Database{
		DB:    client,
		Table: table,
	}
	err = store.Migrate(t.Context())
	xt.NoError(t, err)
	testStore(t, store)
}

func testStore(t *testing.T, store xcookiejar.Storage) {
	jar := &xcookiejar.Jar{
		Storage: store,
	}
	jar = jar.WithContext(t.Context())

	cookies := []*http.Cookie{
		{Name: "name1", Value: "value", Path: "/admin/"},
		{Name: "name2", Value: "value", Path: "/"},
		{Name: "name3", Value: "value", Path: "/"},
	}
	u := &url.URL{Scheme: "http", Host: "example.com", Path: "/"}
	err := jar.SetCookiesContext(t.Context(), u, cookies)
	xt.NoError(t, err)

	items, err := jar.CookiesContext(t.Context(), u)
	xt.NoError(t, err)
	xt.NotEmpty(t, items)
}
