//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
)

// SQLite3 实现 Dialect 接口
type SQLite3 struct{}

// Name 返回方言名称
func (SQLite3) Name() string {
	return "sqlite3"
}

// BindVar 返回占位符。
// SQLite3 支持 "?"，也支持 "?NNN" 形式，但一般直接用 "?"。
func (SQLite3) BindVar(i int) string {
	return "?"
}

// QuoteIdentifier 为标识符添加双引号。
// 注意 SQLite3 中反引号 (`) 也可用，但推荐双引号（兼容 SQL 标准）。
func (SQLite3) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, `"`, `""`)
	return fmt.Sprintf(`"%s"`, safe)
}

// QuoteQualifiedIdentifier 引用多级标识符。
// SQLite3 通常没有 schema，但保留语义。
func (d SQLite3) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

// LimitOffsetClause 生成 LIMIT/OFFSET 子句。
// SQLite3 支持标准写法 "LIMIT ? OFFSET ?"
func (SQLite3) LimitOffsetClause(limit, offset int) string {
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
		// SQLite3 不支持无限 LIMIT，可用极大值替代
		return fmt.Sprintf("LIMIT -1 OFFSET %d", offset)
	}
}

// PlaceholderList 返回 n 个问号占位符
func (SQLite3) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

// SupportsReturning 从 SQLite3 3.35.0 (2021-03) 起支持 RETURNING。
func (SQLite3) SupportsReturning() bool {
	return true
}

// DefaultValueExpr 返回默认值表达式。
func (SQLite3) DefaultValueExpr() string {
	return "DEFAULT"
}

// ReturningClause 生成 RETURNING 子句。
// 空 columns 表示 RETURNING *。
func (SQLite3) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "RETURNING *"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = fmt.Sprintf(`"%s"`, c)
	}
	return "RETURNING " + strings.Join(quoted, ", ")
}

var _ UpsertDialect = SQLite3{}

// UpsertSQL 在 3.24.0 版本（2018年）开始支持
func (d SQLite3) UpsertSQL(table string, count int, columns, conflictCols, updateCols []string, returningCols []string) string {
	colList := strings.Join(xslice.MapFunc(columns, d.QuoteIdentifier), ",")

	valPlaceholders := "(" + strings.Join(xslice.Repeat("?", len(columns)), ",") + ")"

	// INSERT INTO users (id, name, score)
	// VALUES
	//    (1, 'Alice', 100),
	//    (2, 'Bob',   200),
	//    (3, 'Eve',   300)
	// ON CONFLICT(id) DO UPDATE SET
	//    score = excluded.score,
	//    name  = excluded.name;
	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ",
		d.QuoteIdentifier(table),
		colList,
		strings.Join(xslice.Repeat(valPlaceholders, count), ","),
	)

	updateAssignments := make([]string, len(updateCols))
	for i, c := range updateCols {
		c = d.QuoteIdentifier(c)
		updateAssignments[i] = fmt.Sprintf("%s = excluded.%s", c, c)
	}
	sqlStr += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(xslice.MapFunc(conflictCols, d.QuoteIdentifier), ", "),
		strings.Join(updateAssignments, ", "),
	)

	if len(returningCols) > 0 {
		sqlStr += " RETURNING " + strings.Join(returningCols, ", ")
	}
	return sqlStr
}

var _ SchemaDialect = SQLite3{}

func (SQLite3) AutoIncrementColumnType(baseType string, primaryKey bool) string {
	bt := baseType
	if primaryKey {
		baseType += " PRIMARY KEY"
	}
	if bt == "INTEGER" {
		return baseType + " AUTOINCREMENT"
	}
	return baseType
}

func (SQLite3) ColumnType(kind dbcodec.Kind, size int) string {
	switch kind {
	case dbcodec.KindString:
		return "TEXT"
	case dbcodec.KindInt, dbcodec.KindInt8, dbcodec.KindInt16, dbcodec.KindInt32, dbcodec.KindInt64,
		dbcodec.KindUint, dbcodec.KindUint8, dbcodec.KindUint16, dbcodec.KindUint32, dbcodec.KindUint64:
		return "INTEGER"
	case dbcodec.KindBoolean:
		return "INTEGER"
	case dbcodec.KindFloat32, dbcodec.KindFloat64:
		return "REAL"
	case dbcodec.KindJSON:
		// SQLite3 3.38+ 支持 JSON 函数，但底层仍 TEXT
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (d SQLite3) CreateTableIfNotExists(table string) string {
	return "CREATE TABLE IF NOT EXISTS " + d.QuoteIdentifier(table)
}

func (d SQLite3) AddColumnIfNotExists(table string, col string) string {
	// 原生不支持
	return ""
}
