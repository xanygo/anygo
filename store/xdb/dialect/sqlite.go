//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"
	"strings"
)

// SQLite 实现 Dialect 接口
type SQLite struct{}

// Name 返回方言名称
func (SQLite) Name() string {
	return "sqlite"
}

// BindVar 返回占位符。
// SQLite 支持 "?"，也支持 "?NNN" 形式，但一般直接用 "?"。
func (SQLite) BindVar(i int) string {
	return "?"
}

// QuoteIdentifier 为标识符添加双引号。
// 注意 SQLite 中反引号 (`) 也可用，但推荐双引号（兼容 SQL 标准）。
func (SQLite) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, `"`, `""`)
	return fmt.Sprintf(`"%s"`, safe)
}

// QuoteQualifiedIdentifier 引用多级标识符。
// SQLite 通常没有 schema，但保留语义。
func (d SQLite) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

// LimitOffsetClause 生成 LIMIT/OFFSET 子句。
// SQLite 支持标准写法 "LIMIT ? OFFSET ?"
func (SQLite) LimitOffsetClause(limit, offset int) string {
	if limit < 0 && offset <= 0 {
		return ""
	}
	switch {
	case limit >= 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit >= 0:
		return fmt.Sprintf("LIMIT %d", limit)
	default:
		// limit < 0, offset > 0 → “取所有剩余行”
		// SQLite 不支持无限 LIMIT，可用极大值替代
		return fmt.Sprintf("LIMIT -1 OFFSET %d", offset)
	}
}

// PlaceholderList 返回 n 个问号占位符
func (SQLite) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

// SupportsReturning 从 SQLite 3.35.0 (2021-03) 起支持 RETURNING。
func (SQLite) SupportsReturning() bool {
	return true
}

// SupportsUpsert SQLite 从 3.24.0 (2018) 起支持 ON CONFLICT DO UPDATE。
func (SQLite) SupportsUpsert() bool {
	return true
}

// DefaultValueExpr 返回默认值表达式。
func (SQLite) DefaultValueExpr() string {
	return "DEFAULT"
}

// ReturningClause 生成 RETURNING 子句。
// 空 columns 表示 RETURNING *。
func (SQLite) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "RETURNING *"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = fmt.Sprintf(`"%s"`, c)
	}
	return "RETURNING " + strings.Join(quoted, ", ")
}

func (SQLite) UpsertSQL(table string, columns, conflictCols, updateCols []string, args []any, returningCols []string) (string, []any) {
	colList := strings.Join(columns, ", ")

	valPlaceholders := make([]string, len(columns))
	for i := range columns {
		valPlaceholders[i] = "?"
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		colList,
		strings.Join(valPlaceholders, ", "),
	)

	if len(conflictCols) > 0 && len(updateCols) > 0 {
		updateAssignments := make([]string, len(updateCols))
		for i, c := range updateCols {
			updateAssignments[i] = fmt.Sprintf("%s = excluded.%s", c, c)
		}
		sqlStr += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s",
			strings.Join(conflictCols, ", "),
			strings.Join(updateAssignments, ", "),
		)
	}

	if len(returningCols) > 0 {
		sqlStr += " RETURNING " + strings.Join(returningCols, ", ")
	}

	return sqlStr, args
}

func (SQLite) AutoIncrementColumnType(baseType string) string {
	// SQLite 中自增列需定义为 INTEGER PRIMARY KEY AUTOINCREMENT
	if strings.EqualFold(baseType, "integer") {
		return "INTEGER PRIMARY KEY AUTOINCREMENT"
	}
	return baseType
}

func (SQLite) ColumnType(kind string, size int) string {
	switch strings.ToLower(kind) {
	case "string", "varchar":
		if size <= 0 {
			size = 255
		}
		return "TEXT"
	case "int", "integer":
		return "INTEGER"
	case "bigint":
		return "INTEGER"
	case "bool", "boolean":
		return "INTEGER"
	case "float", "double", "real":
		return "REAL"
	case "text":
		return "TEXT"
	case "json":
		// SQLite 3.38+ 支持 JSON 函数，但底层仍 TEXT
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (SQLite) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}
