//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"
	"strings"
)

// Postgres 实现 Dialect 接口
type Postgres struct{}

// Name 返回方言名称
func (Postgres) Name() string {
	return "postgres"
}

// BindVar 返回 PostgreSQL 的占位符 `$1`, `$2`, ...
func (Postgres) BindVar(i int) string {
	return fmt.Sprintf("$%d", i)
}

// QuoteIdentifier 使用双引号包裹标识符
func (Postgres) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, `"`, `""`)
	return fmt.Sprintf(`"%s"`, safe)
}

// QuoteQualifiedIdentifier 支持 schema.table
func (d Postgres) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

// LimitOffsetClause 生成 LIMIT/OFFSET 子句
func (Postgres) LimitOffsetClause(limit, offset int) string {
	switch {
	case limit >= 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit >= 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset > 0:
		// PostgreSQL 不允许 LIMIT ALL OFFSET ?，需明确写 ALL
		return fmt.Sprintf("LIMIT ALL OFFSET %d", offset)
	default:
		return ""
	}
}

// PlaceholderList 返回 n 个占位符 ($1, $2, ...)
func (Postgres) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	holders := make([]string, n)
	for i := 0; i < n; i++ {
		holders[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(holders, ", ")
}

// SupportsReturning 返回 true
func (Postgres) SupportsReturning() bool {
	return true
}

// SupportsUpsert 返回 true
func (Postgres) SupportsUpsert() bool {
	return true
}

// DefaultValueExpr 默认值关键字
func (Postgres) DefaultValueExpr() string {
	return "DEFAULT"
}

// ReturningClause 生成 RETURNING 子句
func (Postgres) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "RETURNING *"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = fmt.Sprintf(`"%s"`, c)
	}
	return "RETURNING " + strings.Join(quoted, ", ")
}

var _ UpsertDialect = (*Postgres)(nil)

func (Postgres) UpsertSQL(table string, cols, conflictCols, updateCols []string, args []any, returningCols []string) (string, []any) {
	colList := strings.Join(cols, ", ")
	valPlaceholders := make([]string, len(cols))
	for i := range cols {
		valPlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}

	updateAssignments := make([]string, len(updateCols))
	for i, c := range updateCols {
		updateAssignments[i] = fmt.Sprintf("%s = EXCLUDED.%s", c, c)
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		table,
		colList,
		strings.Join(valPlaceholders, ", "),
		strings.Join(conflictCols, ", "),
		strings.Join(updateAssignments, ", "),
	)

	if len(returningCols) > 0 {
		sqlStr += " RETURNING " + strings.Join(returningCols, ", ")
	}

	return sqlStr, args
}

// AutoIncrementColumnType PostgreSQL 自增列类型
func (Postgres) AutoIncrementColumnType(baseType string) string {
	switch strings.ToLower(baseType) {
	case "int", "integer":
		return "SERIAL"
	case "bigint":
		return "BIGSERIAL"
	default:
		return baseType
	}
}

// ColumnType 映射通用类型到 PostgreSQL 类型
func (Postgres) ColumnType(kind string, size int) string {
	switch strings.ToLower(kind) {
	case "string", "varchar":
		if size <= 0 {
			return "TEXT"
		}
		return fmt.Sprintf("VARCHAR(%d)", size)
	case "int", "integer":
		return "INTEGER"
	case "bigint":
		return "BIGINT"
	case "bool", "boolean":
		return "BOOLEAN"
	case "float", "double", "real":
		return "DOUBLE PRECISION"
	case "text":
		return "TEXT"
	case "json":
		return "JSONB"
	default:
		return "TEXT"
	}
}

func (Postgres) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}
