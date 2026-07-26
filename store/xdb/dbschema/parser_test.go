//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-17

package dbschema_test

import (
	"slices"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/dialect"
	"github.com/xanygo/anygo/xt"
)

var testUser1Cols = []dbtype.ColumnSchema{
	{
		Name:          "id",
		IsPrimaryKey:  true,
		Kind:          dbtype.KindUint64,
		AutoIncrement: true,
		Codec:         dbcodec.Text{},
	},
	{
		Name:    "name",
		Kind:    dbtype.KindString,
		NotNull: true,
		Unique:  true,
		Codec:   dbcodec.Text{},
	},
	{
		Name:  "roles",
		Kind:  dbtype.KindString,
		Codec: dbcodec.CSV{},
	},
	{
		Name:  "attrs",
		Kind:  dbtype.KindJSON,
		Codec: dbcodec.JSON{},
	},
}

type User1 struct {
	ID    uint64            `db:"id,pk,auto_inc"`
	Name  string            `db:"name,not-null,unique"`
	Roles []int             `db:"roles,codec:csv"`
	Attrs map[string]string `db:"attrs,codec:json"`
}

func TestSchemaUser1(t *testing.T) {
	checkUser1 := func(t *testing.T, sc *dbtype.TableSchema) {
		xt.Empty(t, sc.Table)
		colNames1 := []string{"id", "name", "roles", "attrs"}
		xt.SliceSortEqual(t, colNames1, sc.ColumnsNames)
		for _, col := range testUser1Cols {
			t.Run(col.Name, func(t *testing.T) {
				got, err := sc.ColumnByName(col.Name)
				xt.NoError(t, err)
				xt.NotEmpty(t, got.ReflectType)
				got.ReflectType = nil
				xt.Equal(t, got, col)
			})
		}
		xt.Len(t, sc.Columns, len(testUser1Cols))
	}

	t.Run("struct", func(t *testing.T) {
		sc, err := dbschema.Schema(dialect.MySQL{}, User1{})
		xt.NoError(t, err)
		checkUser1(t, sc)
	})

	t.Run("struct-ptr", func(t *testing.T) {
		sc, err := dbschema.Schema(dialect.MySQL{}, &User1{})
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
	cols = append(cols, dbtype.ColumnSchema{
		Name:  "class",
		Kind:  dbtype.KindString,
		Codec: dbcodec.Text{},
	})
	check := func(t *testing.T, sc *dbtype.TableSchema) {
		xt.Empty(t, sc.Table)
		colNames1 := []string{"id", "name", "roles", "attrs", "class"}
		xt.SliceSortEqual(t, colNames1, sc.ColumnsNames)
		for _, col := range cols {
			t.Run(col.Name, func(t *testing.T) {
				got, err := sc.ColumnByName(col.Name)
				xt.NoError(t, err)
				xt.NotEmpty(t, got.ReflectType)
				got.ReflectType = nil
				xt.Equal(t, got, col)
			})
		}
		xt.Len(t, sc.Columns, len(cols))
	}

	t.Run("struct", func(t *testing.T) {
		sc, err := dbschema.Schema(dialect.MySQL{}, Admin1{})
		xt.NoError(t, err)
		check(t, sc)
	})

	t.Run("struct-ptr", func(t *testing.T) {
		sc, err := dbschema.Schema(dialect.MySQL{}, &Admin1{})
		xt.NoError(t, err)
		check(t, sc)
	})
}

type userTable2 struct {
	ID string `db:"id,pk"`
}

func (ut *userTable2) TableName() string {
	return "ut2"
}

func TestSchemaUser2(t *testing.T) {
	t.Run("case 1 struct", func(t *testing.T) {
		var u userTable2
		sc, err := dbschema.Schema(dialect.MySQL{}, u)
		xt.NoError(t, err)
		xt.Equal(t, sc.Table, "ut2")
	})

	t.Run("case 2 struct ptr", func(t *testing.T) {
		var u *userTable2
		sc, err := dbschema.Schema(dialect.MySQL{}, u)
		xt.NoError(t, err)
		xt.Equal(t, sc.Table, "ut2")
	})
}

type userTable3 struct {
	ID string `db:"id,pk"`
}

func (ut userTable3) TableName() string {
	return "ut3"
}

func TestSchemaUser3(t *testing.T) {
	t.Run("case 1 struct", func(t *testing.T) {
		var u userTable3
		sc, err := dbschema.Schema(dialect.MySQL{}, u)
		xt.NoError(t, err)
		xt.Equal(t, sc.Table, "ut3")
	})

	xt.Panic(t, func() {
		var u *userTable3
		// 目前不支持这样用
		dbschema.Schema(dialect.MySQL{}, u)
	})

	t.Run("case 3 struct ptr", func(t *testing.T) {
		u := &userTable3{}
		sc, err := dbschema.Schema(dialect.MySQL{}, u)
		xt.NoError(t, err)
		xt.Equal(t, sc.Table, "ut3")
	})
}

type userTable4 struct {
	Table     string
	Key       string    `db:"k,unique_index:k_idx"`
	Index     int64     `db:"idx,unique_index:k_idx"`
	Value     string    `db:"v"`
	CreatedAt time.Time `db:"created_at"`
}

func TestSchemaUser4(t *testing.T) {
	u := &userTable4{}
	sc, err := dbschema.Schema(dialect.MySQL{}, u)
	xt.NoError(t, err)
	xt.Equal(t, sc.Table, "")
	// todo:check uniq_index
}
