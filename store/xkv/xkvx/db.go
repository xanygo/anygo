package xkvx

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/store/xkv/internal/db"
)

var _ xkv.StringStorage = (*DatabaseStore)(nil)

type TableProvider struct {
	// 可选，自定义依据 key 获取数据表名
	Resolve func(key string) string

	// 可选，只有在需要 Migrate 的时候才需要
	// 若是定义了 Resolve，若需要  Migrate，则 Names 应为所有可能的表名
	Names []string
}

func (tr *TableProvider) getTable(key string) string {
	if tr == nil || tr.Resolve == nil {
		return ""
	}
	return tr.Resolve(key)
}

func (tr *TableProvider) migrate(ctx context.Context, db xdb.DBCore, obj any, defaultTable string) error {
	if tr == nil || len(tr.Names) == 0 {
		return xor.MigrateWithTable(ctx, db, obj, defaultTable)
	}
	for _, name := range tr.Names {
		if err := xor.MigrateWithTable(ctx, db, obj, name); err != nil {
			return fmt.Errorf("migrate %T (%s):%w", obj, name, err)
		}
	}
	return nil
}

// DatabaseStore 使用数据库存储 KV 类型的数据
//
//	以下是 SQLite 的表结构：
//
//	--- xkv_meta: 存储元信息（所有的 key 以及数据类型）的表
//	--- 下面所有表中的 c 和 u 分别表示数据的创建时间和更新时间，是时间戳(纳秒)
//
//	CREATE TABLE "xkv_meta" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"k" BLOB NOT NULL UNIQUE DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"dt" INTEGER NOT NULL DEFAULT 0,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0,
//	"meta" TEXT NOT NULL DEFAULT ”);
//
//	--- xkv_string：存储 String 类型的数据
//
//	CREATE TABLE "xkv_string" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"k" BLOB NOT NULL UNIQUE DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"v" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0);
//
//	--- xkv_list： 存储 List 类型的数据
//
//	CREATE TABLE "xkv_list" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"idx" INTEGER NOT NULL DEFAULT 0,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"v" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_list_k_i on xkv_list(k,idx);
//
//	---  xkv_hash： 存储 Hash 类型数据
//
//	CREATE TABLE "xkv_hash" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"f" BLOB NOT NULL DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"f_raw" TEXT NOT NULL DEFAULT ”,
//	"v" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_hash_k_f on xkv_hash(k,f);
//
//	--- xkv_set：存储 Set 类型数据
//
//	CREATE TABLE "xkv_set" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"m" BLOB NOT NULL DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"m_raw" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_set_k_m on xkv_set(k,m);
//
//	---  xkv_zset：存储 ZSet 类型数据
//
//	CREATE TABLE "xkv_zset" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"m" BLOB NOT NULL DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"m_raw" TEXT NOT NULL DEFAULT ”,
//	"s" REAL NOT NULL DEFAULT 0,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_zset_k_m on xkv_zset(k,m);
//	CREATE INDEX IF NOT EXISTS idx_xkv_zset_k_s on xkv_zset(k,s);
type DatabaseStore struct {
	// DB 必填字段
	DB *xdb.Client

	// MetaTable 可选，自定义元信息的表名
	MetaTable *TableProvider

	// StringTable 可选，自定义 String 类型数据的表名
	StringTable *TableProvider

	// ListTable 可选，自定义 List 类型数据的表名
	ListTable *TableProvider

	// HashTable 可选，自定义 Hash 类型数据的表名
	HashTable *TableProvider

	// SetTable 可选，自定义 Set 类型数据的表名
	SetTable *TableProvider

	// ZSetTable 可选，自定义 ZSet 类型数据的表名
	ZSetTable *TableProvider
}

func (d *DatabaseStore) String(key string) xkv.String[string] {
	return d.getString(key)
}

func (d *DatabaseStore) getString(key string) *db.String {
	return &db.String{
		Meta:  d.getMeta(key, internal.DataTypeString),
		Table: d.StringTable.getTable(key),
	}
}

func (d *DatabaseStore) getMeta(key string, dt internal.DataType) *db.Meta {
	return &db.Meta{
		Table:    d.MetaTable.getTable(key),
		KeyRaw:   key,
		KeyHash:  db.KeyHash(key),
		DB:       d.DB,
		DataType: dt,
	}
}

func (d *DatabaseStore) List(key string) xkv.List[string] {
	return d.getList(key)
}

func (d *DatabaseStore) getList(key string) *db.List {
	return &db.List{
		Meta:  d.getMeta(key, internal.DataTypeList),
		Table: d.ListTable.getTable(key),
	}
}

func (d *DatabaseStore) Hash(key string) xkv.Hash[string] {
	return d.getHash(key)
}

func (d *DatabaseStore) getHash(key string) *db.Hash {
	return &db.Hash{
		Meta:  d.getMeta(key, internal.DataTypeHash),
		Table: d.HashTable.getTable(key),
	}
}

func (d *DatabaseStore) Set(key string) xkv.Set[string] {
	return d.getSet(key)
}

func (d *DatabaseStore) getSet(key string) *db.Set {
	return &db.Set{
		Meta:  d.getMeta(key, internal.DataTypeSet),
		Table: d.SetTable.getTable(key),
	}
}

func (d *DatabaseStore) ZSet(key string) xkv.ZSet[string] {
	return d.getZSet(key)
}

func (d *DatabaseStore) getZSet(key string) *db.ZSet {
	return &db.ZSet{
		Meta:  d.getMeta(key, internal.DataTypeZSet),
		Table: d.ZSetTable.getTable(key),
	}
}

func (d *DatabaseStore) Has(ctx context.Context, key string) (bool, error) {
	m := d.getMeta(key, internal.DataTypeAny) // 可以是任意类型
	var has bool
	err := m.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		has = hasMeta
		return nil
	})
	return has, err
}

func (d *DatabaseStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	var ms []db.DeleteItem
	for _, key := range keys {
		di := db.DeleteItem{
			Meta:        d.getMeta(key, internal.DataTypeAny),
			StringTable: d.StringTable.getTable(key),
			ListTable:   d.ListTable.getTable(key),
			HashTable:   d.HashTable.getTable(key),
			SetTable:    d.SetTable.getTable(key),
			ZSetTable:   d.ZSetTable.getTable(key),
		}
		ms = append(ms, di)
	}

	dm := db.Delete{
		Items: ms,
	}

	return dm.Delete(ctx)
}

func (d *DatabaseStore) Migrate(ctx context.Context) error {
	metaModel := db.MetaModel{}
	meta := d.getMeta("", internal.DataTypeAny)
	if err := d.MetaTable.migrate(ctx, d.DB, metaModel, meta.GetTable()); err != nil {
		return err
	}
	stringModel := db.StringModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, stringModel, d.getString("").GetTable()); err != nil {
		return err
	}
	listModel := db.ListModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, listModel, d.getList("").GetTable()); err != nil {
		return err
	}
	hashModel := db.HashModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, hashModel, d.getHash("").GetTable()); err != nil {
		return err
	}
	setModel := db.SetModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, setModel, d.getSet("").GetTable()); err != nil {
		return err
	}
	zsetModel := db.ZSetModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, zsetModel, d.getZSet("").GetTable()); err != nil {
		return err
	}
	return nil
}
