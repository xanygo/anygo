package xkvx

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/store/xkv/internal"
	"github.com/xanygo/anygo/store/xkv/internal/db"
)

var _ xkv.StringStorage = (*Database)(nil)

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

// Init 用于从配置中初始化
func (tr *TableProvider) Init(param map[string]any) error {
	name, _ := xmap.GetString(param, "Name")
	total, _ := xmap.GetInt(param, "Total")
	if name == "" {
		return nil
	}
	if total <= 1 {
		tr.Names = []string{name}
		tr.Resolve = func(key string) string {
			return name
		}
		return nil
	}

	for i := 0; i < total; i++ {
		tr.Names = append(tr.Names, fmt.Sprintf("%s_%d", name, i))
	}
	tr.Resolve = func(key string) string {
		h := fnv.New32a()
		h.Write([]byte(key))
		index := int(h.Sum32()) % len(tr.Names)
		return tr.Names[index]
	}
	return nil
}

// Database 使用数据库存储 KV 类型的数据
//
//	以下是 SQLite 的表结构：
//
//	--- xkv_meta: 存储元信息（所有的 key 以及数据类型）的表
//	--- 下面所有表中的 c 和 u 分别表示数据的创建时间和更新时间，是时间戳(纳秒)
//
//	CREATE TABLE "xkv_meta" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"t" INTEGER NOT NULL DEFAULT 0,
//	"k" BLOB NOT NULL UNIQUE DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"dt" INTEGER NOT NULL DEFAULT 0,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0,
//	"meta" TEXT NOT NULL DEFAULT ”);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_meta_t_k on xkv_meta(t,k);
//
//	--- xkv_string：存储 String 类型的数据
//
//	CREATE TABLE "xkv_string" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"t" INTEGER NOT NULL DEFAULT 0,
//	"k" BLOB NOT NULL DEFAULT '',
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"v" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_string_t_k on xkv_string(t,k);
//
//	--- xkv_list： 存储 List 类型的数据
//
//	CREATE TABLE "xkv_list" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"t" INTEGER NOT NULL DEFAULT 0,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"idx" INTEGER NOT NULL DEFAULT 0,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"v" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_list_t_k_i on xkv_list(t,k,idx);
//
//	---  xkv_hash： 存储 Hash 类型数据
//
//	CREATE TABLE "xkv_hash" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"t" INTEGER NOT NULL DEFAULT 0,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"f" BLOB NOT NULL DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"f_raw" TEXT NOT NULL DEFAULT ”,
//	"v" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_hash_t_k_f on xkv_hash(t,k,f);
//
//	--- xkv_set：存储 Set 类型数据
//
//	CREATE TABLE "xkv_set" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"t" INTEGER NOT NULL DEFAULT 0,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"m" BLOB NOT NULL DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"m_raw" TEXT NOT NULL DEFAULT ”,
//	"c" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_set_t_k_m on xkv_set(t,k,m);
//
//	---  xkv_zset：存储 ZSet 类型数据
//
//	CREATE TABLE "xkv_zset" (
//	"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//	"t" INTEGER NOT NULL DEFAULT 0,
//	"k" BLOB NOT NULL DEFAULT ”,
//	"m" BLOB NOT NULL DEFAULT ”,
//	"k_raw" TEXT NOT NULL DEFAULT ”,
//	"m_raw" TEXT NOT NULL DEFAULT ”,
//	"s" REAL NOT NULL DEFAULT 0,
//	"c" INTEGER NOT NULL DEFAULT 0,
//	"u" INTEGER NOT NULL DEFAULT 0);
//
//	CREATE UNIQUE INDEX IF NOT EXISTS idx_xkv_zset_t_k_m on xkv_zset(t,k,m);
//	CREATE INDEX IF NOT EXISTS idx_xkv_zset_t_k_s on xkv_zset(t,k,s);
type Database struct {
	// DB 必填字段
	DB *xdb.Client

	ValueTypeID uint32 // 可选，存储数据的实际类型签名,可以使用 GenTypeID 获取

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

var stringTypeID = zreflect.TypeID1[string]()

func (d *Database) getTypeID() uint32 {
	if d.ValueTypeID > 0 {
		return d.ValueTypeID
	}
	return stringTypeID
}

func (d *Database) GenTypeID[V any]() uint32 {
	return zreflect.TypeID1[V]()
}

func (d *Database) Init(param map[string]any) error {
	if d.DB == nil {
		service, ok := xmap.GetString(param, "Service")
		if !ok || service == "" {
			return fmt.Errorf("no Service in %v", param)
		}
		db, err := xdb.NewClientWithService(service)
		if err != nil {
			return err
		}
		d.DB = db
	}
	var err error
	if d.MetaTable == nil {
		d.MetaTable, err = d.initTableProvider(param, "MetaTable")
	}

	if err == nil && d.StringTable == nil {
		d.StringTable, err = d.initTableProvider(param, "StringTable")
	}
	if err == nil && d.ListTable == nil {
		d.ListTable, err = d.initTableProvider(param, "ListTable")
	}
	if err == nil && d.HashTable == nil {
		d.HashTable, err = d.initTableProvider(param, "HashTable")
	}
	if err == nil && d.SetTable == nil {
		d.SetTable, err = d.initTableProvider(param, "SetTable")
	}
	if err == nil && d.ZSetTable == nil {
		d.ZSetTable, err = d.initTableProvider(param, "ZSetTable")
	}
	return err
}

func (d *Database) initTableProvider(param map[string]any, name string) (*TableProvider, error) {
	cfg, ok := xmap.GetMap(param, name)
	if !ok || len(cfg) == 0 {
		return nil, nil
	}
	tp := &TableProvider{}
	if err := tp.Init(cfg); err != nil {
		return nil, err
	}
	return tp, nil
}

func (d *Database) String(key string) xkv.String[string] {
	return d.getString(key)
}

func (d *Database) getString(key string) *db.String {
	return &db.String{
		Meta:  d.getMeta(key, internal.DataTypeString),
		Table: d.StringTable.getTable(key),
	}
}

func (d *Database) getMeta(key string, dt internal.DataType) *db.Meta {
	return &db.Meta{
		TypeID:   d.getTypeID(),
		Table:    d.MetaTable.getTable(key),
		KeyRaw:   key,
		KeyHash:  db.KeyHash(key),
		DB:       d.DB,
		DataType: dt,
	}
}

func (d *Database) List(key string) xkv.List[string] {
	return d.getList(key)
}

func (d *Database) getList(key string) *db.List {
	return &db.List{
		Meta:  d.getMeta(key, internal.DataTypeList),
		Table: d.ListTable.getTable(key),
	}
}

func (d *Database) Hash(key string) xkv.Hash[string] {
	return d.getHash(key)
}

func (d *Database) getHash(key string) *db.Hash {
	return &db.Hash{
		Meta:  d.getMeta(key, internal.DataTypeHash),
		Table: d.HashTable.getTable(key),
	}
}

func (d *Database) Set(key string) xkv.Set[string] {
	return d.getSet(key)
}

func (d *Database) getSet(key string) *db.Set {
	return &db.Set{
		Meta:  d.getMeta(key, internal.DataTypeSet),
		Table: d.SetTable.getTable(key),
	}
}

func (d *Database) ZSet(key string) xkv.ZSet[string] {
	return d.getZSet(key)
}

func (d *Database) getZSet(key string) *db.ZSet {
	return &db.ZSet{
		Meta:  d.getMeta(key, internal.DataTypeZSet),
		Table: d.ZSetTable.getTable(key),
	}
}

func (d *Database) Has(ctx context.Context, key string) (bool, error) {
	m := d.getMeta(key, internal.DataTypeAny) // 可以是任意类型
	var has bool
	err := m.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		has = hasMeta
		return nil
	})
	return has, err
}

func (d *Database) Delete(ctx context.Context, keys ...string) error {
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

func (d *Database) Migrate(ctx context.Context) error {
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
