//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

type Dialect interface {
	// Name 返回方言名称，如 "postgres", "mysql", "sqlite3"
	Name() string

	// BindVar 返回第 i 个绑定变量的占位符。
	// 例如：Postgres -> "$1", MySQL/SQLite -> "?"
	// i 从 1 开始。
	BindVar(i int) string

	// QuoteIdentifier 将一个标识符（表名、列名）引用起来（包括 schema 或带点的部分应正确处理）。
	// 例如：QuoteIdentifier("user") -> `"user"` 或 `` `user` ``。
	QuoteIdentifier(s string) string

	// QuoteQualifiedIdentifier 对于带 schema 的标识符可能有特殊处理：
	// e.g. QuoteQualifiedIdentifier("public", "user") -> `"public"."user"`
	// 如果不需要可以用默认实现（库层提供 helper）。
	QuoteQualifiedIdentifier(parts ...string) string

	// LimitOffsetClause 返回给定 limit/offset 的 SQL 片段（不包含前后的空格）
	// 举例："LIMIT 10 OFFSET 20" 或 "LIMIT 20, 10"（MySQL）。
	// 当 limit<0 且 offset<=0 时返回 ""。
	LimitOffsetClause(limit, offset int) string

	// PlaceholderList 返回 n 个绑定占位符组成的字符串，用于 IN (...) 或批量插入。
	// start 表示占位符起始序号（便于 Postgres 系列生成 $1,$2...）。
	// 例如：PlaceholderList(3, 1) -> "$1,$2,$3" 或 "?,?,?"。
	PlaceholderList(n, start int) string

	// SupportsReturning 表示方言是否支持 INSERT ... RETURNING / UPDATE ... RETURNING
	SupportsReturning() bool

	// SupportsUpsert 表示是否有原生 upsert（ON CONFLICT / ON DUPLICATE KEY UPDATE 等）
	SupportsUpsert() bool

	// DefaultValueExpr 返回 DB 的默认值表达式（例如 "DEFAULT" 或 ""）
	DefaultValueExpr() string
}

// ReturningDialect 提供 RETURNING 子句生成（仅对支持的方言实现）
type ReturningDialect interface {
	// ReturningClause 返回 RETURNING 子句（不包含前导空格），columns 为空表示返回所有列（若方言支持）。
	ReturningClause(columns ...string) string
}

// UpsertDialect 提供 upsert 片段生成
type UpsertDialect interface {
	// UpsertSQL
	// table: 表名
	// cols: 所有字段
	// conflictCols: 冲突字段（主键或唯一键）
	// updateCols: 冲突时需要更新的字段
	// args: 对应参数值
	// returningCols: 可选返回字段
	// 返回可执行 SQL + 参数切片
	UpsertSQL(table string, cols, conflictCols, updateCols []string, args []any, returningCols []string) (string, []any)
}

// SchemaDialect DDL 相关扩展（创建表、列类型等）
type SchemaDialect interface {
	// AutoIncrementColumnType 返回用于自增的列类型或后缀，如 "SERIAL" / "INTEGER PRIMARY KEY AUTOINCREMENT"。
	AutoIncrementColumnType(baseType string) string

	// ColumnType 将通用列类型映射为方言列类型（例如 "string" -> "VARCHAR(255)"）。
	ColumnType(kind string, size int) string

	// CreateTableIfNotExists 返回 CREATE TABLE ... IF NOT EXISTS 的片段（或空串如果不支持）。
	CreateTableIfNotExists() string
}

// JSONDialect JSON 操作相关（Postgres JSONB / MySQL JSON）
type JSONDialect interface {
	// JSONExtractExpr 返回从 jsonCol 中取出路径 path 的表达式，例如 "data->'a'->>0" 或 "JSON_EXTRACT(data, '$.a[0]')"
	JSONExtractExpr(jsonCol string, path []string) string
}

// Capability Feature detection：也可通过 Capability 方法扩展
type Capability interface {
	// HasFeature 返回方言是否支持某个特性标识（例如 "window_functions", "cte", "json"）
	HasFeature(feature string) bool
}
