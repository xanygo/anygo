package xkvx

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/store/xkv/internal/db"
)

var _ xkv.StringStorage = (*DatabaseStorage)(nil)

type TableProvider struct {
	// 可选，自定义依据 key 获取数据表名
	Resolve func(key string) string

	// 可选，只有在需要 Migrate 的时候才需要
	Names []string
}

func (tr *TableProvider) getTable(key string) string {
	if tr == nil || tr.Resolve == nil {
		return ""
	}
	return tr.Resolve(key)
}

func (tr *TableProvider) migrate(ctx context.Context, db xdb.DBCore, obj any) error {
	if tr == nil || len(tr.Names) == 0 {
		return xdb.MigrateWithTable(ctx, db, obj, "")
	}
	for _, name := range tr.Names {
		if err := xdb.MigrateWithTable(ctx, db, obj, name); err != nil {
			return fmt.Errorf("migrate %T (%s):%w", obj, name, err)
		}
	}
	return nil
}

// DatabaseStorage 使用数据库存储 KV 类型的数据
//
//	以下是 SQLite 的表结构：
//
// --- xkv_meta: 存储元信息（所有的 key 以及数据类型）的表
// --- 下面所有表中的 c 和 u 分别表示数据的创建时间和更新时间，是 unix 时间戳
// CREATE TABLE IF NOT EXISTS xkv_meta (k TEXT PRIMARY KEY,dt INTEGER,meta TEXT,c INTEGER,u INTEGER);
//
// --- xkv_string：存储 String 类型的数据
// CREATE TABLE IF NOT EXISTS xkv_string (k TEXT PRIMARY KEY,v TEXT,c INTEGER,u INTEGER);
//
// --- xkv_list： 存储 List 类型的数据
// CREATE TABLE IF NOT EXISTS xkv_list (k TEXT,idx INTEGER,v TEXT,c INTEGER);
// CREATE UNIQUE INDEX IF NOT EXISTS idx_k_i on xkv_list(k,idx);
//
// ---  xkv_hash： 存储 Hash 类型数据
// CREATE TABLE IF NOT EXISTS xkv_hash (k TEXT,f TEXT,v TEXT,c INTEGER,u INTEGER);
// CREATE UNIQUE INDEX IF NOT EXISTS idx_k_f on xkv_hash(k,f);
//
// --- xkv_set：存储 Set 类型数据
// CREATE TABLE IF NOT EXISTS xkv_set (k TEXT,m TEXT,c INTEGER);
// CREATE UNIQUE INDEX IF NOT EXISTS idx_k_m on xkv_set(k,m);
//
// ---  xkv_zset：存储 ZSet 类型数据
// CREATE TABLE IF NOT EXISTS xkv_zset (k TEXT,m TEXT,s REAL,c INTEGER,u INTEGER);
// CREATE INDEX IF NOT EXISTS idx_k_i on xkv_zset(k,s);
// CREATE UNIQUE INDEX IF NOT EXISTS idx_k_m on xkv_zset(k,m);
type DatabaseStorage struct {
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

func (d *DatabaseStorage) String(key string) xkv.String[string] {
	return &db.String{
		Meta: db.MetaModel{
			Table:    d.MetaTable.getTable(key),
			Key:      key,
			DB:       d.DB,
			DataType: internal.DataTypeString,
		},
		Table: d.StringTable.getTable(key),
		Key:   key,
	}
}

func (d *DatabaseStorage) List(key string) xkv.List[string] {
	return &db.List{
		Meta: db.MetaModel{
			Table:    d.MetaTable.getTable(key),
			Key:      key,
			DB:       d.DB,
			DataType: internal.DataTypeList,
		},
		Table: d.ListTable.getTable(key),
		Key:   key,
	}
}

func (d *DatabaseStorage) Hash(key string) xkv.Hash[string] {
	return &db.Hash{
		Meta: db.MetaModel{
			Table:    d.MetaTable.getTable(key),
			Key:      key,
			DB:       d.DB,
			DataType: internal.DataTypeHash,
		},
		Table: d.HashTable.getTable(key),
		Key:   key,
	}
}

func (d *DatabaseStorage) Set(key string) xkv.Set[string] {
	return &db.Set{
		Meta: db.MetaModel{
			Table:    d.MetaTable.getTable(key),
			Key:      key,
			DB:       d.DB,
			DataType: internal.DataTypeSet,
		},
		Table: d.SetTable.getTable(key),
		Key:   key,
	}
}

func (d *DatabaseStorage) ZSet(key string) xkv.ZSet[string] {
	return &db.ZSet{
		Meta: db.MetaModel{
			Table:    d.MetaTable.getTable(key),
			Key:      key,
			DB:       d.DB,
			DataType: internal.DataTypeZSet,
		},
		Table: d.ZSetTable.getTable(key),
		Key:   key,
	}
}

func (d *DatabaseStorage) Has(ctx context.Context, key string) (bool, error) {
	m := db.MetaModel{
		Table:    d.MetaTable.getTable(key),
		Key:      key,
		DB:       d.DB,
		DataType: internal.DataTypeAny, // 可以是任意类型
	}
	var has bool
	err := m.WithReadTx(ctx, func(ctx context.Context, tx xdb.TxCore, hasMeta bool) error {
		has = hasMeta
		return nil
	})
	return has, err
}

func (d *DatabaseStorage) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	var ms []db.DeleteItem
	for _, key := range keys {
		di := db.DeleteItem{
			Meta: db.MetaModel{
				Table: d.MetaTable.getTable(key),
				Key:   key,
				DB:    d.DB,
			},
		}
		ms = append(ms, di)
	}

	dm := db.Delete{
		Items: ms,
	}

	return dm.Delete(ctx)
}

func (d *DatabaseStorage) Migrate(ctx context.Context) error {
	meta := db.MetaModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, meta); err != nil {
		return err
	}
	stringModel := db.StringModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, stringModel); err != nil {
		return err
	}
	listModel := db.ListModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, listModel); err != nil {
		return err
	}
	hashModel := db.HashModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, hashModel); err != nil {
		return err
	}
	setModel := db.SetModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, setModel); err != nil {
		return err
	}
	zsetModel := db.ZSetModel{}
	if err := d.MetaTable.migrate(ctx, d.DB, zsetModel); err != nil {
		return err
	}
	return nil
}
