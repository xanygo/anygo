//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb_test

import (
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xtdr"
	"github.com/xanygo/anygo/xt"
)

func TestScanRows(t *testing.T) {
	db := xtdr.MustOpen()
	t.Run("case 1", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name"}, [][]any{{1, "hello"}, {2, "world"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser1](ret)
		xt.NoError(t, err)
		want := []testUser1{
			{ID: 1, Name: "hello"},
			{ID: 2, Name: "world"},
		}
		xt.Equal(t, want, users)
	})

	t.Run("case 2", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name"}, [][]any{{1, "hello"}, {2, "world"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[map[string]any](ret)
		xt.NoError(t, err)
		want := []map[string]any{
			{"id": int64(1), "name": "hello"},
			{"id": int64(2), "name": "world"},
		}
		xt.Equal(t, want, users)
	})

	t.Run("case 3", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"ids1", "Bs1"}, [][]any{{"[1]", "hello"}, {"", "world"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser1](ret)
		xt.NoError(t, err)
		want := []testUser1{
			{IDs1: []int{1}, Bs1: []byte("hello")},
			{IDs1: nil, Bs1: []byte("world")},
		}
		xt.Equal(t, want, users)
	})

	t.Run("case 4", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"csv1", "mp1"}, [][]any{{"1,2", `{"k1":"v1"}`}, {"", ""}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser2](ret)
		xt.NoError(t, err)
		want := []testUser2{
			{CSV1: []int{1, 2}, MP1: map[string]string{"k1": "v1"}},
			{CSV1: nil, MP1: nil},
		}
		xt.Equal(t, want, users)
	})
}

type testUser2 struct {
	CSV1 []int             `db:"csv1,codec:csv"`
	MP1  map[string]string `db:"mp1,codec:json"`
}
