//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-10

package xdb

import (
	"context"
	"errors"
	"fmt"
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
	dialect       dialect.Dialect
	client        HasDriver
	table         string
	limit, offset int
	err           error
}

func (m *Model[T]) Reset() *Model[T] {
	m.limit = 0
	m.offset = 0
	return m
}

func (m *Model[T]) Clone() *Model[T] {
	return &Model[T]{
		dialect: m.dialect,
		client:  m.client,
		table:   m.table,
		limit:   m.limit,
		offset:  m.offset,
		err:     m.err,
	}
}

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

// Insert 基本的 Insert 功能
func (m *Model[T]) Insert(ctx context.Context, v T) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	kv, err := Encode(v)
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
	values, err := EncodeBatch[T](vs...)
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
	if m.err != nil {
		return 0, m.err
	}

	kv, err := Encode(v)
	if err != nil {
		return 0, err
	}
	cols := xmap.Keys(kv)

	where, args, err = m.buildWhere(where, args)
	if err != nil {
		return 0, err
	}
	if len(where) == 0 || len(args) == 0 {
		return 0, errors.New("empty where clause")
	}
	assigns := make([]string, len(cols))
	for i, col := range cols {
		assigns[i] = fmt.Sprintf(`%s=%s`, m.dialect.QuoteIdentifier(col), m.dialect.BindVar(i+1))
	}

	sqlStr := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		m.dialect.QuoteIdentifier(m.table),
		strings.Join(assigns, ", "),
		where,
	)
	vals := xmap.Values(kv)
	vals = append(vals, args...)

	db, ok := m.client.(Execer)
	if !ok {
		return 0, fmt.Errorf("client (%T) is not Execer", m.client)
	}
	ret, err := Exec(ctx, db, sqlStr, vals...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

func (m *Model[T]) Delete(ctx context.Context, where string, args ...any) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	var err error
	where, args, err = m.buildWhere(where, args)
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

func (m *Model[T]) First(ctx context.Context, fields []string, where string, args ...any) (v T, ok bool, err error) {
	if m.err != nil {
		return v, false, m.err
	}
	where, args, err = m.buildWhere(where, args)
	if err != nil {
		return v, false, err
	}
	var field string
	if len(fields) == 0 {
		field = "*"
	} else {
		field = strings.Join(xslice.MapFunc(fields, m.dialect.QuoteIdentifier), ",")
	}
	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s %s",
		field,
		m.dialect.QuoteIdentifier(m.table),
		where,
		m.dialect.LimitOffsetClause(1, 0),
	)
	db, ok := m.client.(Queryer)
	if !ok {
		return v, false, fmt.Errorf("client (%T) is not Queryer", m.client)
	}
	return QueryOne[T](ctx, db, sqlStr, args...)
}

func (m *Model[T]) List(ctx context.Context, fields []string, where string, args ...any) ([]T, error) {
	if m.err != nil {
		return nil, m.err
	}
	var field string
	if len(fields) == 0 {
		field = "*"
	} else {
		field = strings.Join(xslice.MapFunc(fields, m.dialect.QuoteIdentifier), ",")
	}
	w1, a1, err := m.buildWhere(where, args)
	if err != nil {
		return nil, err
	}
	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		field,
		m.dialect.QuoteIdentifier(m.table),
		w1,
	)
	db, ok := m.client.(Queryer)
	if !ok {
		return nil, fmt.Errorf("client (%T) is not Queryer", m.client)
	}
	return QueryMany[T](ctx, db, sqlStr, a1...)
}

func (m *Model[T]) buildWhere(where string, args []any) (string, []any, error) {
	if m.dialect.BindVar(0) == "?" {
		return where, args, nil
	}

	// 将 ? 替换为 $1, $2 ...
	var sb strings.Builder
	idx := 1
	for i := 0; i < len(where); i++ {
		if where[i] == '?' {
			sb.WriteString(m.dialect.BindVar(idx))
			idx++
		} else {
			sb.WriteByte(where[i])
		}
	}
	return sb.String(), args, nil
}
