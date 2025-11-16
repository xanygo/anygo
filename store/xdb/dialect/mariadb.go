//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"context"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbschema"
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
	return (MySQL{}).LimitOffsetClause(limit, offset)
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

var _ UpsertDialect = MariaDB{}

func (d MariaDB) UpsertSQL(table string, count int, columns, conflictCols, updateCols []string, returningCols []string) string {
	colList := strings.Join(xslice.MapFunc(columns, d.QuoteIdentifier), ",")

	valPlaceholders := "(" + strings.Join(xslice.Repeat("?", len(columns)), ",") + ")"

	updateAssignments := make([]string, len(updateCols))
	for i, c := range updateCols {
		c = d.QuoteIdentifier(c)
		updateAssignments[i] = fmt.Sprintf("%s = VALUES(%s)", c, c)
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		table,
		colList,
		strings.Join(xslice.Repeat(valPlaceholders, count), ","),
		strings.Join(updateAssignments, ", "),
	)

	return sqlStr
}

var _ SchemaDialect = MariaDB{}

func (MariaDB) ColumnType(kind dbcodec.Kind, size int) string {
	return (MySQL{}).ColumnType(kind, size)
}

func (d MariaDB) ColumnString(fs *dbschema.ColumnSchema) string {
	return (MySQL{}).ColumnString(fs)
}

func (d MariaDB) CreateTableIfNotExists(table string) string {
	return "CREATE TABLE IF NOT EXISTS " + d.QuoteIdentifier(table)
}

//	func (d MariaDB) addColumnIfNotExists(table string, col string) string {
//		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", d.QuoteIdentifier(table), d.QuoteIdentifier(col))
//	}
var _ MigrateDialect = MariaDB{}

func (d MariaDB) Migrate(ctx context.Context, db DBCore, schema dbschema.TableSchema) error {
	sqlStr := createTableSQL(schema, d, d)
	_, err := db.ExecContext(ctx, sqlStr)
	return err
}
