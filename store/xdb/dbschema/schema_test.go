//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-17

package dbschema_test

import (
	"slices"
	"testing"

	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/xt"
)

var testUser1Cols = []*dbschema.ColumnSchema{
	{
		Name:          "id",
		IsPrimaryKey:  true,
		Kind:          dbcodec.KindUint64,
		AutoIncrement: true,
	},
	{
		Name:    "name",
		Kind:    dbcodec.KindString,
		NotNull: true,
		Unique:  true,
	},
	{
		Name:  "roles",
		Kind:  dbcodec.KindString,
		Codec: "csv",
	},
	{
		Name:  "attrs",
		Kind:  dbcodec.KindJSON,
		Codec: "json",
	},
}

type User1 struct {
	ID    uint64            `db:"id,pk,autoInc"`
	Name  string            `db:"name,notNull,unique"`
	Roles []int             `db:"roles,codec:csv"`
	Attrs map[string]string `db:"attrs,codec:json"`
}

func TestSchemaUser1(t *testing.T) {
	checkUser1 := func(t *testing.T, sc *dbschema.TableSchema) {
		xt.Empty(t, sc.Table)
		colNames1 := []string{"id", "name", "roles", "attrs"}
		xt.SliceSortEqual(t, colNames1, sc.ColumnsNames())
		for _, col := range testUser1Cols {
			t.Run(col.Name, func(t *testing.T) {
				got, err := sc.ColumnByName(col.Name)
				xt.NoError(t, err)
				xt.Equal(t, col, got)
			})
		}
		xt.Len(t, sc.Columns, len(testUser1Cols))
	}

	t.Run("struct", func(t *testing.T) {
		sc, err := dbschema.Schema(User1{})
		xt.NoError(t, err)
		checkUser1(t, sc)
	})

	t.Run("struct-ptr", func(t *testing.T) {
		sc, err := dbschema.Schema(&User1{})
		xt.NoError(t, err)
		checkUser1(t, sc)
	})
}

type Admin1 struct {
	User1
	Class  string `db:"class"`
	Other1 string // 会被忽略
	Other2 string `db:"-"` // 会被忽略
}

func TestSchemaAdmin1(t *testing.T) {
	cols := slices.Clone(testUser1Cols)
	cols = append(cols, &dbschema.ColumnSchema{
		Name: "class",
		Kind: dbcodec.KindString,
	})
	check := func(t *testing.T, sc *dbschema.TableSchema) {
		xt.Empty(t, sc.Table)
		colNames1 := []string{"id", "name", "roles", "attrs", "class"}
		xt.SliceSortEqual(t, colNames1, sc.ColumnsNames())
		for _, col := range cols {
			t.Run(col.Name, func(t *testing.T) {
				got, err := sc.ColumnByName(col.Name)
				xt.NoError(t, err)
				xt.Equal(t, col, got)
			})
		}
		xt.Len(t, sc.Columns, len(cols))
	}

	t.Run("struct", func(t *testing.T) {
		sc, err := dbschema.Schema(Admin1{})
		xt.NoError(t, err)
		check(t, sc)
	})

	t.Run("struct-ptr", func(t *testing.T) {
		sc, err := dbschema.Schema(&Admin1{})
		xt.NoError(t, err)
		check(t, sc)
	})
}
