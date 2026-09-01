//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xanygo/anygo/safely"
)

type DBCore interface {
	Queryer
	Execer
	RowQuerier
}

type (
	// Queryer 封装执行查询并返回多行结果的方法
	Queryer interface {
		HasDriver
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}

	StmtQueryer interface {
		HasDriver
		QueryContext(ctx context.Context, args ...any) (*sql.Rows, error)
	}
)

// Execer 封装执行 SQL 语句的方法
type (
	Execer interface {
		HasDriver
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	}

	StmtExecer interface {
		HasDriver
		ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	}
)

type (
	// RowQuerier 封装执行查询并返回单行结果的方法
	RowQuerier interface {
		HasDriver
		QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	}

	StmtRowQuerier interface {
		HasDriver
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
	return ScanRows[T](q, rows)
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
		for k, v := range ScanRowsIter[T](q, rows) {
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
	return ScanRowsFirst[T](q, rows)
}

// Exec 执行写语句(insert、update、delete)
func Exec(ctx context.Context, eq Execer, query string, args ...any) (ret sql.Result, err error) {
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

// BeginTx 支持嵌套的开启事务(已对 do 添加 panic recover 处理)
//
// 可以在 do 方法里调用 BeginTx 开启子事务(采用的 SavePoint 方式)
func BeginTx(ctx context.Context, core DBCore, opts *sql.TxOptions, do func(ctx context.Context, tx DBCore) error) error {
	switch client := core.(type) {
	case canBeginTx:
		te, err := client.BeginTx(ctx, opts)
		if err != nil {
			return err
		}
		return WithTx(ctx, te, do)
	case canSavePoint:
		return startTxSavePoint(ctx, client, do)
	}
	return fmt.Errorf("not supporttype %T with StartTx", core)
}

// WithTx 在事务中安全的处理逻辑(已对 do 添加 panic recover 处理)
//
//	若 do 方法 返回 nil，会自动执行 Commit
//	若 do 方法 返回 error 或者 panic，会自动的执行 Rollback。
func WithTx(ctx context.Context, tx TxExecutor, do func(ctx context.Context, tx DBCore) error) error {
	err := safely.RunCtx(ctx, func(ctx context.Context) error {
		return do(ctx, tx)
	})
	if err == nil {
		return tx.Commit()
	}
	err1 := tx.Rollback()
	return errors.Join(err, err1)
}

var savePointID atomic.Int64

func startTxSavePoint(ctx context.Context, tx canSavePoint, do func(ctx context.Context, tx DBCore) error) error {
	name := fmt.Sprintf("sp_%d", savePointID.Add(1))
	if err := tx.SavePoint(ctx, name); err != nil {
		return err
	}
	err := safely.RunCtx(ctx, func(ctx context.Context) error {
		return do(ctx, tx)
	})
	if err != nil {
		err1 := tx.RollbackTo(ctx, name)
		return errors.Join(err, err1)
	}
	return tx.ReleaseSavepoint(ctx, name)
}

func StmtQueryMany[T any](ctx context.Context, q Statement, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return ScanRows[T](q, rows)
}

func StmtQueryOne[T any](ctx context.Context, q Statement, args ...any) (v T, ok bool, err error) {
	rows, err := q.QueryContext(ctx, args...)
	if err != nil {
		return v, false, err
	}
	return ScanRowsFirst[T](q, rows)
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
		for k, v := range ScanRowsIter[T](q, rows) {
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

var _ error = (*QueryError)(nil)

func newQueryError(err error, caller string, query string, args []any) *QueryError {
	return &QueryError{
		Caller: caller,
		Query:  query,
		Args:   args,
		Raw:    err,
	}
}

type QueryError struct {
	Caller string
	Query  string
	Args   []any
	Raw    error

	str  string
	once sync.Once
}

func (q *QueryError) Error() string {
	q.once.Do(func() {
		var txt string
		if q.Raw != nil {
			txt = q.Raw.Error()
		}
		var sb strings.Builder
		sb.WriteString(q.Caller)
		sb.WriteString(" ")
		sb.WriteString(txt)
		sb.WriteString(", query=")
		fmt.Fprintf(&sb, "%q", q.Query)
		fmt.Fprintf(&sb, ", args_len=%d,", len(q.Args))
		for i, a := range q.Args {
			fmt.Fprintf(&sb, ", [%d](%T)=%v", i, a, a)
		}
		q.str = sb.String()
	})
	return q.str
}

func (q *QueryError) Unwrap() error {
	return q.Raw
}
