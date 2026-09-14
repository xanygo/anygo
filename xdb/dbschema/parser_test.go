//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-17

package dbschema_test

import (
	"slices"
	"testing"
	"time"
	"uuid"

	"github.com/xanygo/anygo/xdb/dbcodec"
	"github.com/xanygo/anygo/xdb/dbschema"
	"github.com/xanygo/anygo/xdb/dbtype"
	"github.com/xanygo/anygo/xdb/dialect"
	"github.com/xanygo/anygo/xslice"
	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xtime"
)

var testUser1Cols = []dbtype.ColumnSchema{
	{
		Name:          "id",
		IsPrimaryKey:  true,
		Kind:          dbtype.KindUint64,
		AutoIncrement: true,
		Codec:         &dbcodec.Native{},
		NotNull:       true,
		Group:         []string{},
	},
	{
		Name:    "name",
		Kind:    dbtype.KindString,
		NotNull: true,
		Unique:  true,
		Codec:   &dbcodec.Text{},
		Group:   []string{"a"},
	},
	{
		Name:  "roles",
		Kind:  dbtype.KindString,
		Codec: &dbcodec.CSV{},
		Group: []string{},
	},
	{
		Name:    "attrs",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
	{
		Name:    "sign",
		Kind:    dbtype.KindBinary,
		NotNull: true,
		Codec:   &dbcodec.Binary{},
		Size:    32,
		Group:   []string{},
	},
	{
		Name:    "uuid",
		Kind:    dbtype.KindUUID,
		NotNull: true,
		Codec:   &dbcodec.UUID{},
		Size:    16,
		Group:   []string{},
	},
	{
		Name:    "i8",
		Kind:    dbtype.KindInt8,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "i16",
		Kind:    dbtype.KindInt16,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "i32",
		Kind:    dbtype.KindInt32,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "i64",
		Kind:    dbtype.KindInt64,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "u8",
		Kind:    dbtype.KindUint8,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "u16",
		Kind:    dbtype.KindUint16,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "u32",
		Kind:    dbtype.KindUint32,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "uint",
		Kind:    dbtype.KindUint,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "created",
		Kind:    dbtype.KindMilliseconds,
		NotNull: true,
		Codec:   &dbcodec.Milliseconds{},
		Group:   []string{},
		Auto:    "Created",
	},
	{
		Name:    "updated",
		Kind:    dbtype.KindMilliseconds,
		NotNull: true,
		Codec:   &dbcodec.Milliseconds{},
		Group:   []string{},
		Auto:    "Updated",
	},
	{
		Name:    "version",
		Kind:    dbtype.KindInt,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
		Auto:    "Incr",
	},
	{
		Name:    "meta",
		Kind:    dbtype.KindBinary,
		NotNull: true,
		Codec:   &dbcodec.Binary{},
		Group:   []string{},
	},
	{
		Name:    "dur1",
		Kind:    dbtype.KindInt64,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "dur2",
		Kind:    dbtype.KindInt64,
		NotNull: true,
		Codec:   &dbcodec.Native{},
		Group:   []string{},
	},
	{
		Name:    "struct1",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
	{
		Name:    "ptr1",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
	{
		Name:    "map1",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
	{
		Name:    "map2",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
	{
		Name:    "slice1",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
	{
		Name:    "slice2",
		Kind:    dbtype.KindJSON,
		NotNull: true,
		Codec:   &dbcodec.JSON{},
		Group:   []string{},
	},
}

type User1 struct {
	ID      uint64            `db:"id,pk,auto_inc"`
	Name    string            `db:"name,not-null,unique,group=a"`
	Roles   []int             `db:"roles,codec=csv,null"`
	Attrs   map[string]string `db:"attrs,codec=json"`
	Sign    [32]byte          `db:"sign"`
	UUID    uuid.UUID         `db:"uuid"`
	I8      int8              `db:"i8"`
	I16     int16             `db:"i16"`
	I32     int32             `db:"i32"`
	I64     int64             `db:"i64"`
	U8      uint8             `db:"u8"`
	U16     uint16            `db:"u16"`
	U32     uint32            `db:"u32"`
	Uint    uint              `db:"uint"`
	Created time.Time         `db:"created,auto=Created"`
	Updated time.Time         `db:"updated,auto=Updated"`
	Version int               `db:"version,auto=Incr"`
	Meta    []byte            `db:"meta"`
	Dur1    time.Duration     `db:"dur1"`
	Dur2    xtime.Duration    `db:"dur2"`
	Struct1 Struct1           `db:"struct1"`
	Ptr1    *Struct1          `db:"ptr1"`
	Map1    map[string]any    `db:"map1"`
	Map2    map[string]int    `db:"map2"`
	Slice1  []uuid.UUID       `db:"slice1"`
	Slice2  []Struct1         `db:"slice2"`
}

type Struct1 struct {
	Name string
}

func TestSchemaUser1(t *testing.T) {
	checkUser1 := func(t *testing.T, sc *dbtype.TableSchema) {
		xt.Empty(t, sc.Table)
		colNames1 := xslice.FilterAs(testUser1Cols, func(item dbtype.ColumnSchema) (string, bool) {
			return item.Name, true
		})
		xt.SliceSortEqual(t, colNames1, sc.ColumnNames)
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
		Name:    "class",
		Kind:    dbtype.KindString,
		NotNull: true,
		Codec:   &dbcodec.Text{},
		Group:   []string{},
	})
	check := func(t *testing.T, sc *dbtype.TableSchema) {
		xt.Empty(t, sc.Table)
		colNames1 := xslice.FilterAs(testUser1Cols, func(item dbtype.ColumnSchema) (string, bool) {
			return item.Name, true
		})
		colNames1 = append(colNames1, "class")
		xt.SliceSortEqual(t, colNames1, sc.ColumnNames)
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

	xt.NoPanic(t, func() {
		var u *userTable3
		// 目前不支持这样用
		dbschema.Schema(dialect.MySQL{}, u)
	})

	t.Run("case 3 nil ptr", func(t *testing.T) {
		var u *userTable3
		sc, err := dbschema.Schema(dialect.MySQL{}, u)
		xt.NoError(t, err)
		xt.Equal(t, sc.Table, "ut3")
		xt.Equal(t, sc.ColumnNames, []string{"id"})
	})

	t.Run("case 4 struct ptr", func(t *testing.T) {
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

func BenchmarkSchema(b *testing.B) {
	md := dialect.MySQL{}
	u := &userTable4{}
	for i := 0; i < b.N; i++ {
		_, err := dbschema.Schema(md, u)
		xt.NoError(b, err)
	}
}
