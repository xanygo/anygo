//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-08-28

package xor

import (
	"slices"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/dialect"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
)

// HasTable 给 Model 使用的 struct 可以选择实现该接口，以自动读取数据库表名
type HasTable interface {
	TableName() string
}

// NewMode 生成一个 T 类型的 简单 ORM Model
// T 类型建议传入 struct 而不是 *struct，兼容性更好
func New[T any](db xdb.DBCore) *Model[T] {
	m := &Model[T]{
		client: db,
	}
	m.init()
	return m
}

// Model 轻量 ORM 实现，已实现数据模型常用的增删改查功能
type Model[T any] struct {
	dialect dbtype.Dialect
	client  xdb.DBCore

	schema *dbtype.TableSchema  // 不会为空
	table  string               // 表名
	pk     dbtype.ColumnSchemas // 可能为 nil

	err error // 初始化过程中的错误
	cfg *config
}

func (m *Model[T]) init() {
	m.dialect, m.err = dialect.Find(m.client.Driver())
	if m.err == nil {
		var zero T
		m.schema, m.err = dbschema.Schema(m.dialect, zero)
	}
	if m.schema != nil {
		m.table = m.schema.Table
		m.pk = m.schema.PKColumns()
	}
	m.cfg = &config{
		dialect: m.dialect,
		schema:  m.schema,
	}
}

func (m *Model[T]) New() *Model[T] {
	return &Model[T]{
		dialect: m.dialect,
		client:  m.client,
		table:   m.table,
		pk:      slices.Clone(m.pk),
		schema:  m.schema,
		err:     m.err,
		cfg: &config{
			schema:  m.schema,
			dialect: m.dialect,
		},
	}
}

func (m *Model[T]) Clone() *Model[T] {
	return &Model[T]{
		dialect: m.dialect,
		client:  m.client,
		table:   m.table,
		pk:      slices.Clone(m.pk),
		schema:  m.schema,
		err:     m.err,
		cfg:     m.cfg.Clone(),
	}
}

func (m *Model[T]) DB() xdb.DBCore {
	return m.client
}

func (m *Model[T]) With(opts ...Option) {
	m.cfg.merge(opts...)
}

// Table 设置表名，若 T 没有实现 HasTable 接口时，可通过此设置
func (m *Model[T]) Table(table string) *Model[T] {
	m.table = table
	return m
}

func (m *Model[T]) getEncoder(action encoder.Action, cfg *config) encoder.Encoder[T] {
	return encoder.Encoder[T]{
		Schema:       m.schema,
		Action:       action,
		Dialect:      m.dialect,
		OnlyFields:   cfg.getColumns(),
		IgnoreFields: cfg.getIgnores(),
	}
}

// Quote 将标识符转义
func (m *Model[T]) Quote(name string) string {
	return m.dialect.QuoteIdentifier(name)
}
