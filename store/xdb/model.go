//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-10

package xdb

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb/dialect"
)

type HasTable interface {
	TableName() string
}

func NewMode[T any](client HasDriver) *Model[T] {
	m := &Model[T]{
		client: client,
	}
	m.dialect, m.err = dialect.Find(client.Driver())
	return m
}

type Model[T any] struct {
	dialect dialect.Dialect
	client  HasDriver

	table         string
	limit, offset int
	onlyFields    []string // insert, update 的字段列表
	ignoreFields  []string // select, update 时候忽略的字段列表

	err error
}

func (m *Model[T]) Reset() *Model[T] {
	m.limit = 0
	m.offset = 0
	m.onlyFields = nil
	m.ignoreFields = nil
	return m
}

func (m *Model[T]) Clone() *Model[T] {
	return &Model[T]{
		dialect:      m.dialect,
		client:       m.client,
		table:        m.table,
		limit:        m.limit,
		offset:       m.offset,
		err:          m.err,
		onlyFields:   slices.Clone(m.onlyFields),
		ignoreFields: slices.Clone(m.ignoreFields),
	}
}

// OnlyFields 设置 insert、update 的字段列表，默认为空时，写入所有字段
func (m *Model[T]) OnlyFields(fields ...string) *Model[T] {
	m.onlyFields = fields
	return m
}

func (m *Model[T]) AppendOnlyFields(fields ...string) *Model[T] {
	m.onlyFields = append(m.onlyFields, fields...)
	return m
}

// IgnoreFields 设置 select 和 update 时候，需要忽略的字段，默认为空
func (m *Model[T]) IgnoreFields(fields ...string) *Model[T] {
	m.ignoreFields = fields
	return m
}

func (m *Model[T]) AppendIgnoreFields(fields ...string) *Model[T] {
	m.ignoreFields = append(m.ignoreFields, fields...)
	return m
}

// Table 设置表名，若 T 没有实现 HasTable 接口时，可通过此设置
func (m *Model[T]) Table(table string) *Model[T] {
	m.table = table
	return m
}

func (m *Model[T]) Limit(num int) *Model[T] {
	m.limit = num
	return m
}

func (m *Model[T]) Offset(num int) *Model[T] {
	m.offset = num
	return m
}

func (m *Model[T]) getEncoder() Encoder[T] {
	return Encoder[T]{
		OnlyFields:   m.onlyFields,
		IgnoreFields: m.ignoreFields,
	}
}

// Insert 基本的 Insert 功能
func (m *Model[T]) Insert(ctx context.Context, v T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	kv, err := m.getEncoder().Encode(v)
	if err != nil {
		return 0, err
	}
	if len(kv) == 0 {
		return 0, errors.New("no columns")
	}

	qcols := xslice.MapFunc(xmap.Keys(kv), m.dialect.QuoteIdentifier)

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(qcols, ", "),
		m.dialect.PlaceholderList(len(kv), 1),
	)
	if r, ok := any(m.dialect).(dialect.ReturningDialect); ok {
		sqlStr += " " + r.ReturningClause()
	}

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, xmap.Values(kv)...)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

func (m *Model[T]) InsertBatch(ctx context.Context, vs ...T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if len(vs) == 0 {
		return 0, errors.New("no values")
	}
	values, err := m.getEncoder().EncodeBatch(vs...)
	if err != nil {
		return 0, err
	}
	cols := xmap.Keys(values[0])
	if len(cols) == 0 {
		return 0, errors.New("no columns")
	}

	qCols := xslice.MapFunc(cols, m.dialect.QuoteIdentifier)
	holder := "(" + m.dialect.PlaceholderList(len(cols), 1) + ")"
	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(qCols, ","),
		strings.Join(xslice.Repeat(holder, len(values)), ", "),
	)
	if r, ok := any(m.dialect).(dialect.ReturningDialect); ok {
		sqlStr += " " + r.ReturningClause()
	}

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}

	vals := make([]any, 0, len(values)*len(cols))
	for _, item := range values {
		vals = append(vals, xmap.Values(item)...)
	}

	ret, err := Exec(ctx, db, sqlStr, vals...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

func (m *Model[T]) Update(ctx context.Context, v T, where string, args ...any) (int64, error) {
	return m.doUpdate(ctx, v, where, args...)
}

func (m *Model[T]) doUpdate(ctx context.Context, v T, where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	kv, err := m.getEncoder().Encode(v)
	if err != nil {
		return 0, err
	}
	cols := xmap.Keys(kv)

	assigns := make([]string, 0, len(cols))
	values := make([]any, 0, len(args))
	for _, col := range cols {
		str := fmt.Sprintf(`%s=%s`, m.dialect.QuoteIdentifier(col), m.dialect.BindVar(len(assigns)+1))
		assigns = append(assigns, str)
		values = append(values, kv[col])
	}

	if len(assigns) == 0 {
		return 0, errors.New("no update values")
	}

	where, args, err = m.buildWhere(len(assigns), where, args)
	if err != nil {
		return 0, err
	}
	if len(where) == 0 || len(args) == 0 {
		return 0, errors.New("empty where clause")
	}

	sqlStr := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(assigns, ", "),
		where,
	)
	values = append(values, args...)

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, values...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// UpdateByPK 使用主键更新数据
//
// 需要再 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,primaryKey"`
func (m *Model[T]) UpdateByPK(ctx context.Context, v T) (int64, error) {
	pk, value, err := findStructPrimaryKV(v)
	if err != nil {
		return 0, err
	}
	where := m.dialect.QuoteIdentifier(pk) + "=?"

	m1 := m.Clone()
	m1.AppendIgnoreFields(pk)
	return m1.doUpdate(ctx, v, where, value)
}

func (m *Model[T]) Delete(ctx context.Context, where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	var err error
	where, args, err = m.buildWhere(0, where, args)
	if err != nil {
		return 0, err
	}
	if len(where) == 0 || len(args) == 0 {
		return 0, errors.New("empty where clause")
	}
	sqlStr := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		m.dialect.QuoteIdentifier(m.table),
		where,
	)
	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// DeleteByPK 使用主键删除数据
//
// 需要再 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,primaryKey"`
func (m *Model[T]) DeleteByPK(ctx context.Context, v T) (int64, error) {
	pk, value, err := findStructPrimaryKV(v)
	if err != nil {
		return 0, err
	}
	where := m.dialect.QuoteIdentifier(pk) + "=?"
	return m.Delete(ctx, where, value)
}

func (m *Model[T]) selectFields() (string, error) {
	var zero T
	fields, err := m.getEncoder().Fields(zero)
	if err != nil {
		return "", err
	}
	if len(fields) == 0 {
		return "*", nil
	}
	return strings.Join(xslice.MapFunc(fields, m.dialect.QuoteIdentifier), ","), nil
}

func (m *Model[T]) First(ctx context.Context, where string, args ...any) (v T, ok bool, err error) {
	if m.err != nil {
		return v, false, m.err
	}
	where, args, err = m.buildWhere(0, where, args)
	if err != nil {
		return v, false, err
	}
	field, err := m.selectFields()
	if err != nil {
		return v, false, err
	}
	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s %s %s",
		field,
		m.dialect.QuoteIdentifier(m.table),
		m.connectWhere(where),
		m.dialect.LimitOffsetClause(1, 0),
	)
	db, ok := m.client.(Queryer)
	if !ok {
		return v, false, fmt.Errorf("client (%T) is not Queryer", m.client)
	}
	return QueryOne[T](ctx, db, sqlStr, args...)
}

// FindByPK 使用主键查找数据
//
// 需要再 tag 里有 primaryKey 属性: 如 ID int64 `db:"id,primaryKey"`
func (m *Model[T]) FindByPK(ctx context.Context, v T) (nv T, ok bool, err error) {
	pk, value, err := findStructPrimaryKV(v)
	if err != nil {
		return nv, false, err
	}
	where := m.dialect.QuoteIdentifier(pk) + "=?"
	return m.First(ctx, where, value)
}

func (m *Model[T]) List(ctx context.Context, where string, args ...any) ([]T, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []T
	for item, err := range m.ListIter(ctx, where, args...) {
		if err != nil {
			return result, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (m *Model[T]) ListIter(ctx context.Context, where string, args ...any) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		if m.err != nil {
			yield(zero, m.err)
			return
		}

		field, err := m.selectFields()
		if err != nil {
			yield(zero, err)
			return
		}

		where, args, err = m.buildWhere(0, where, args)
		if err != nil {
			yield(zero, err)
			return
		}

		sqlStr := fmt.Sprintf(
			"SELECT %s FROM %s %s",
			field,
			m.dialect.QuoteIdentifier(m.table),
			m.connectWhere(where),
		)
		db, ok := m.client.(Queryer)
		if !ok {
			err = fmt.Errorf("client (%T) is not Queryer", m.client)
			yield(zero, err)
			return
		}
		for item, err := range QueryManyIter[T](ctx, db, sqlStr, args...) {
			if !yield(item, err) {
				return
			}
		}
	}
}

func (m *Model[T]) connectWhere(where string) string {
	if where == "" {
		return ""
	}
	return " where " + where
}

func (m *Model[T]) buildWhere(indexStart int, where string, args []any) (string, []any, error) {
	if m.dialect.BindVar(0) == "?" {
		return where, args, nil
	}

	// 将 ? 替换为 $1, $2 ...
	var sb strings.Builder
	idx := 1
	for i := 0; i < len(where); i++ {
		if where[i] == '?' {
			sb.WriteString(m.dialect.BindVar(indexStart + idx))
			idx++
		} else {
			sb.WriteByte(where[i])
		}
	}
	return sb.String(), args, nil
}
