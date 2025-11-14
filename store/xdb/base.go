//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb

import (
	"context"
	"database/sql"
	"iter"

	"github.com/xanygo/anygo/safely"
)

type DBCore interface {
	Queryer
	Execer
}

type (
	// Queryer 封装执行查询并返回多行结果的方法
	Queryer interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}

	StmtQueryer interface {
		QueryContext(ctx context.Context, args ...any) (*sql.Rows, error)
	}
)

// Execer 封装执行 SQL 语句的方法
type (
	Execer interface {
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	}

	StmtExecer interface {
		ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	}
)

type (
	// RowQuerier 封装执行查询并返回单行结果的方法
	RowQuerier interface {
		QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	}

	StmtRowQuerier interface {
		QueryRowContext(ctx context.Context, args ...any) *sql.Row
	}
)

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
	StmtQueryer
	StmtExecer
	StmtRowQuerier
	Close() error
}

// QueryMany 执行查询 SQL，并返回匹配的全部结果
func QueryMany[T any](ctx context.Context, q Queryer, query string, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return ScanRows[T](rows)
}

// QueryManyIter 行查询 SQL，并返回匹配结果的迭代器。只有读取完，或者退出迭代器，底层链接才会是否
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

// QueryOne 查询并返回匹配的首条结果
func QueryOne[T any](ctx context.Context, q Queryer, query string, args ...any) (v T, ok bool, err error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return v, false, err
	}
	return ScanRowsFirst[T](rows)
}

// Exec 执行写语句(insert、update、delete)
func Exec(ctx context.Context, eq Execer, query string, args ...any) (sql.Result, error) {
	return eq.ExecContext(ctx, query, args...)
}

func LastInsertID(ret sql.Result, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

func RowsAffected(ret sql.Result, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
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

func StmtQueryManyIter[T any](ctx context.Context, q Statement, args ...any) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		rows, err := q.QueryContext(ctx, args...)
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

func StmtExec(ctx context.Context, eq Statement, args ...any) (sql.Result, error) {
	return eq.ExecContext(ctx, args...)
}

func Count(ctx context.Context, q RowQuerier, query string, args ...any) (num int64, err error) {
	row := q.QueryRowContext(ctx, query, args...)
	err = row.Scan(&num)
	return num, err
}
