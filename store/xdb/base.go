//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb

import (
	"context"
	"database/sql"
	"iter"
	"log"

	"github.com/xanygo/anygo/safely"
)

// Execer 封装执行 SQL 语句的方法
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Queryer 封装执行查询并返回多行结果的方法
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// RowQuerier 封装执行查询并返回单行结果的方法
type RowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Preparer 封装预编译语句的方法
type Preparer interface {
	PrepareContext(ctx context.Context, query string) (Statement, error)
}

type TxExecutor interface {
	TxCore
	Commit() error
	Rollback() error
}

type TxCore interface {
	Queryer
	Execer
	Preparer
	RowQuerier
	StmtContext(ctx context.Context, stmt Statement) Statement
}

type Statement interface {
	Unwrap() *sql.Stmt
	QueryContext(ctx context.Context, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, args ...any) *sql.Row
	Close() error
}

func QueryMany[T any](ctx context.Context, q Queryer, query string, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return ScanRows[T](rows)
}

func QueryManyIter[T any](ctx context.Context, q Queryer, query string, args ...any) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		rows, err := q.QueryContext(ctx, query, args...)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		defer rows.Close()
		for k, v := range ScanRowsIter[T](rows) {
			yield(k, v)
		}
	}
}

func QueryOne[T any](ctx context.Context, q Queryer, query string, args ...any) (v T, ok bool, err error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return v, false, err
	}
	return ScanRowsFirst[T](rows)
}

func Exec(ctx context.Context, eq Execer, query string, args ...any) (sql.Result, error) {
	log.Println("Exec:", query, args)
	return eq.ExecContext(ctx, query, args...)
}

func WithTx(ctx context.Context, tx TxExecutor, do func(ctx context.Context, tx TxCore) error) error {
	err := safely.RunCtx(ctx, func(ctx context.Context) error {
		return do(ctx, tx)
	})
	if err == nil {
		return tx.Commit()
	}
	return tx.Rollback()
}

func StmtQueryMany[T any](ctx context.Context, q Statement, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return ScanRows[T](rows)
}

func StmtQueryOne[T any](ctx context.Context, q Statement, query string, args ...any) (v T, ok bool, err error) {
	rows, err := q.QueryContext(ctx, args...)
	if err != nil {
		return v, false, err
	}
	return ScanRowsFirst[T](rows)
}

var _ TxExecutor = (*tx)(nil)
var _ HasDriver = (*tx)(nil)

type tx struct {
	Raw    *sql.Tx
	driver string
}

func (t *tx) Driver() string {
	return t.driver
}

func (t *tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.Raw.QueryContext(ctx, query, args...)
}

func (t *tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.Raw.ExecContext(ctx, query, args...)
}

func (t *tx) PrepareContext(ctx context.Context, query string) (Statement, error) {
	s, err := t.Raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &stmt{Raw: s}, nil
}

func (t *tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.Raw.QueryRowContext(ctx, query, args...)
}

func (t *tx) StmtContext(ctx context.Context, s Statement) Statement {
	st := t.Raw.StmtContext(ctx, s.Unwrap())
	return &stmt{Raw: st}
}

func (t *tx) Commit() error {
	return t.Raw.Commit()
}

func (t *tx) Rollback() error {
	return t.Raw.Rollback()
}

var _ Statement = (*stmt)(nil)
var _ HasDriver = (*stmt)(nil)

type stmt struct {
	Raw    *sql.Stmt
	driver string
}

func (s *stmt) Driver() string {
	return s.driver
}

func (s *stmt) Unwrap() *sql.Stmt {
	return s.Raw
}

func (s *stmt) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	return s.Raw.QueryContext(ctx, args...)
}

func (s *stmt) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	return s.Raw.ExecContext(ctx, args...)
}

func (s *stmt) QueryRowContext(ctx context.Context, args ...any) *sql.Row {
	return s.Raw.QueryRowContext(ctx, args...)
}

func (s *stmt) Close() error {
	return s.Raw.Close()
}
