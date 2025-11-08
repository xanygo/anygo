//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb

import (
	"context"
	"database/sql"
)

// Executor 封装执行 SQL 语句的方法
type Executor interface {
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
}

// Querier 封装执行查询并返回多行结果的方法
type Querier interface {
	QueryContext(ctx context.Context, args ...any) (*sql.Rows, error)
}

// RowQuerier 封装执行查询并返回单行结果的方法
type RowQuerier interface {
	QueryRowContext(ctx context.Context, args ...any) *sql.Row
}

// Preparer 封装预编译语句的方法
type Preparer interface {
	PrepareContext(ctx context.Context, query string) (Statement, error)
}

type TxExecutor interface {
	Querier
	Executor
	Preparer
	RowQuerier

	StmtContext(ctx context.Context, stmt Statement) Statement

	Commit() error
	Rollback() error
}

type Statement interface {
	Querier
	Executor
	RowQuerier
	Close() error
}
