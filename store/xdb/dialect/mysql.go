//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"
	"strings"
)

var _ Dialect = (*MySQL)(nil)

type MySQL struct{}

// Name 返回方言名称
func (MySQL) Name() string {
	return "mysql"
}

// BindVar 返回绑定变量占位符。
// MySQL 使用 "?" 占位符，忽略序号。
func (MySQL) BindVar(i int) string {
	return "?"
}

// QuoteIdentifier 为标识符添加反引号。
// 若标识符中包含反引号，则替换为双反引号（避免语法错误）。
func (MySQL) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, "`", "``")
	return fmt.Sprintf("`%s`", safe)
}

// QuoteQualifiedIdentifier 引用多级标识符（例如 schema.table.column）
func (d MySQL) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

// LimitOffsetClause 生成 LIMIT/OFFSET 语句片段。
// MySQL 支持两种写法：
//
//	LIMIT 10 OFFSET 20
//	LIMIT 20,10
//
// 通常推荐使用前者（兼容性更好）
func (MySQL) LimitOffsetClause(limit, offset int) string {
	if limit < 0 && offset <= 0 {
		return ""
	}
	switch {
	case limit >= 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit >= 0:
		return fmt.Sprintf("LIMIT %d", limit)
	default:
		// limit < 0, offset > 0 不常见
		// 效果上等价于 “无限大 LIMIT”,当于告诉 MySQL “跳过前 offset 条，然后返回后面所有剩余的行”。
		// 2^64 - 1 -> unsigned BIGINT 的最大值
		return fmt.Sprintf("LIMIT 18446744073709551615 OFFSET %d", offset)
	}
}

// PlaceholderList 返回 n 个问号占位符，用逗号分隔。
func (MySQL) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

// SupportsReturning MySQL 直到 8.0.19 仍不支持 RETURNING 子句（除非用 MariaDB）。
func (MySQL) SupportsReturning() bool {
	return false
}

// SupportsUpsert MySQL 支持 ON DUPLICATE KEY UPDATE。
func (MySQL) SupportsUpsert() bool {
	return true
}

// DefaultValueExpr MySQL 默认值关键字
func (MySQL) DefaultValueExpr() string {
	return "DEFAULT"
}

func (MySQL) UpsertSQL(table string, columns, conflictCols, updateCols []string, args []any, returningCols []string) (string, []any) {
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

func (MySQL) AutoIncrementColumnType(baseType string) string {
	// 例如 baseType="BIGINT" => "BIGINT AUTO_INCREMENT"
	return baseType + " AUTO_INCREMENT"
}

func (MySQL) ColumnType(kind string, size int) string {
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
	default:
		return kind
	}
}

func (MySQL) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}
