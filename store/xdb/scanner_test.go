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

type testUser1 struct {
	sid    int
	ID     int               `db:"id"`
	Name   string            `db:"name"`
	Enable bool              `db:"enable"`
	Score  float64           `db:"score"`
	IDs1   []int             `db:"ids1,codec=json"`
	IDs2   []int             `db:"ids2,codec=json"`
	IDs3   []int             // 没有定义 db tag，会被忽略
	Md1    map[string]any    `db:"md1,codec=json"`
	Md2    map[string]string `db:"md2,codec=json"`

	Bs1 []byte `db:"bs1"`
	ID2 *int   `db:"id2"`
}

var _ = &testUser1{sid: 1}

type TestUser2 struct {
	CSV1 []int             `db:"csv1,codec=csv"`
	MP1  map[string]string `db:"mp1,codec=json"`
}

type TestUser21 struct {
	U21 string `db:"u21"`
	TestUser2
}

type TestUser22 struct {
	// 多层嵌套
	U22 string `db:"u22"`

	TestUser21
}

type testUser3 struct {
	Name      string `db:"name"`
	skip      string
	TestUser2 ``
}

type testUser4 struct {
	Name       string `db:"name"`
	skip       string
	*TestUser2 // 在 scanner 中，这个要求是可导出类型的
}

var _ = testUser4{skip: "ok"}

type testUser5 struct {
	CSV1 []int             `db:"csv1,codec=csv"`
	MP1  map[string]string `db:"mp1,codec=json"`
}

type testUser6 struct {
	Name       string `db:"name"`
	skip       string
	*testUser5 // 这个是不可导出类型，所以不能通过反射设置值
}

var _ = testUser6{skip: "ok"}

func TestScanRows(t *testing.T) {
	db := xtdr.MustOpen()
	client := xdb.NewClient("mysql", "test", db)
	defer xtdr.Reset()
	t.Run("case 1", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name"}, [][]any{{1, "hello"}, {2, "world"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser1](client, ret)
		xt.NoError(t, err)
		want := []testUser1{
			{ID: 1, Name: "hello"},
			{ID: 2, Name: "world"},
		}
		xt.Equal(t, users, want)
	})

	t.Run("case 2", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name"}, [][]any{{1, "hello"}, {2, "world"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[map[string]any](client, ret)
		xt.NoError(t, err)
		want := []map[string]any{
			{"id": int64(1), "name": "hello"},
			{"id": int64(2), "name": "world"},
		}
		xt.Equal(t, users, want)
	})

	t.Run("case 3", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"ids1", "bs1"}, [][]any{{"[1]", "hello"}, {"", "world"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser1](client, ret)
		xt.NoError(t, err)
		want := []testUser1{
			{IDs1: []int{1}, Bs1: []byte("hello")},
			{IDs1: nil, Bs1: []byte("world")},
		}
		xt.Equal(t, users, want)
	})

	t.Run("case 4", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"csv1", "mp1"}, [][]any{{"1,2", `{"k1":"v1"}`}, {"", ""}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[TestUser2](client, ret)
		xt.NoError(t, err)
		want := []TestUser2{
			{CSV1: []int{1, 2}, MP1: map[string]string{"k1": "v1"}},
			{CSV1: nil, MP1: nil},
		}
		xt.Equal(t, users, want)
	})
}

func TestScanRowsEmbed(t *testing.T) {
	db := xtdr.MustOpen()
	client := xdb.NewClient("pgx", "test", db)

	t.Run("testUser3", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name", "csv1"}, [][]any{{1, "hello", "1,2"}, {2, "world", "2,3"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser3](client, ret)
		xt.NoError(t, err)
		want := []testUser3{
			{Name: "hello", TestUser2: TestUser2{CSV1: []int{1, 2}}, skip: ""},
			{Name: "world", TestUser2: TestUser2{CSV1: []int{2, 3}}},
		}
		xt.Equal(t, users, want)
	})

	t.Run("testUser4", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name", "csv1"}, [][]any{{1, "hello", "1,2"}, {2, "world", "2,3"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser4](client, ret)
		xt.NoError(t, err)
		want := []testUser4{
			{Name: "hello", TestUser2: &TestUser2{CSV1: []int{1, 2}}},
			{Name: "world", TestUser2: &TestUser2{CSV1: []int{2, 3}}},
		}
		xt.Equal(t, users, want)
	})
	t.Run("testUser6", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name", "csv1"}, [][]any{{1, "hello", "1,2"}, {2, "world", "2,3"}})
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[testUser6](client, ret)
		xt.Error(t, err)
		xt.Empty(t, users)
		xt.ErrorContains(t, err, "Cannot Set")
	})

	t.Run("testUser22", func(t *testing.T) {
		values := [][]any{{"u22-value", "u21-hello", "1,2", ""}, {"hello", "world", "2,3", `{"k1":"v1"}`}}
		xtdr.ExpectQuery("select 1", []string{"u22", "u21", "csv1", "mp1"}, values)
		ret, err := db.Query("select 1")
		xt.NoError(t, err)
		users, err := xdb.ScanRows[TestUser22](client, ret)
		xt.NoError(t, err)
		want := []TestUser22{
			{
				U22: "u22-value",
				TestUser21: TestUser21{
					U21: "u21-hello",
					TestUser2: TestUser2{
						CSV1: []int{1, 2},
					},
				},
			},
			{
				U22: "hello",
				TestUser21: TestUser21{
					U21: "world",
					TestUser2: TestUser2{
						CSV1: []int{2, 3},
						MP1: map[string]string{
							"k1": "v1",
						},
					},
				},
			},
		}
		xt.Equal(t, users, want)
	})
}
