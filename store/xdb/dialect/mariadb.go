//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"
	"strings"
)

// MariaDB 实现 Dialect 接口
type MariaDB struct{}

// Name 返回方言名称
func (MariaDB) Name() string {
	return "mariadb"
}

// BindVar 返回绑定变量占位符。
// MariaDB 与 MySQL 一致，使用 "?"。
func (MariaDB) BindVar(i int) string {
	return "?"
}

// QuoteIdentifier 为标识符添加反引号。
// 如果内部有反引号，则替换为双反引号。
func (MariaDB) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, "`", "``")
	return fmt.Sprintf("`%s`", safe)
}

// QuoteQualifiedIdentifier 引用多级标识符（如 schema.table）
func (d MariaDB) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

// LimitOffsetClause 生成 LIMIT/OFFSET 语句。
// 与 MySQL 一致。
func (MariaDB) LimitOffsetClause(limit, offset int) string {
	if limit < 0 && offset <= 0 {
		return ""
	}
	switch {
	case limit >= 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit >= 0:
		return fmt.Sprintf("LIMIT %d", limit)
	default:
		// limit < 0, offset > 0 → 表示“取所有剩余行”
		return fmt.Sprintf("LIMIT 18446744073709551615 OFFSET %d", offset)
	}
}

// PlaceholderList 返回 n 个问号占位符，用逗号分隔。
func (MariaDB) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

// SupportsReturning MariaDB 从 10.5 开始支持 RETURNING。
func (MariaDB) SupportsReturning() bool {
	return true
}

// SupportsUpsert MariaDB 支持 ON DUPLICATE KEY UPDATE。
func (MariaDB) SupportsUpsert() bool {
	return true
}

// DefaultValueExpr 默认值关键字
func (MariaDB) DefaultValueExpr() string {
	return "DEFAULT"
}

// ReturningClause 返回 RETURNING 子句。
// 如果 columns 为空，返回 RETURNING *。
func (MariaDB) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "RETURNING *"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = fmt.Sprintf("`%s`", c)
	}
	return "RETURNING " + strings.Join(quoted, ", ")
}

func (MariaDB) UpsertSQL(table string, columns, conflictCols, updateCols []string, args []any, returningCols []string) (string, []any) {
	colList := strings.Join(columns, ", ")
	valPlaceholders := make([]string, len(columns))
	for i := range columns {
		valPlaceholders[i] = "?"
	}

	updateAssignments := make([]string, len(updateCols))
	for i, c := range updateCols {
		updateAssignments[i] = fmt.Sprintf("%s = VALUES(%s)", c, c)
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		table,
		colList,
		strings.Join(valPlaceholders, ", "),
		strings.Join(updateAssignments, ", "),
	)

	return sqlStr, args
}

func (MariaDB) AutoIncrementColumnType(baseType string) string {
	return baseType + " AUTO_INCREMENT"
}

func (MariaDB) ColumnType(kind string, size int) string {
	switch strings.ToLower(kind) {
	case "string", "varchar":
		if size <= 0 {
			size = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", size)
	case "int", "integer":
		return "INT"
	case "bigint":
		return "BIGINT"
	case "bool", "boolean":
		return "TINYINT(1)"
	case "float":
		return "FLOAT"
	case "double":
		return "DOUBLE"
	case "text":
		return "TEXT"
	case "json":
		// MariaDB 没有真正 JSON 类型，用 LONGTEXT 模拟
		return "LONGTEXT"
	default:
		return kind
	}
}

func (MariaDB) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}
