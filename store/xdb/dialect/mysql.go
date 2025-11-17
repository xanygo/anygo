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

var _ UpsertDialect = MySQL{}

func (d MySQL) UpsertSQL(table string, count int, columns, conflictCols, updateCols []string, returningCols []string) string {
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

var _ SchemaDialect = MySQL{}

func (MySQL) ColumnType(kind dbcodec.Kind, size int) string {
	switch kind {
	case dbcodec.KindString:
		if size <= 0 {
			size = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", size)

	case dbcodec.KindInt:
		return "INT"
	case dbcodec.KindInt8:
		return "TINYINT"
	case dbcodec.KindInt16:
		return "SMALLINT"
	case dbcodec.KindInt32:
		return "INT"
	case dbcodec.KindInt64:
		return "BIGINT"

	case dbcodec.KindUint:
		return "INT UNSIGNED"
	case dbcodec.KindUint8:
		return "TINYINT UNSIGNED"
	case dbcodec.KindUint16:
		return "SMALLINT UNSIGNED"
	case dbcodec.KindUint32:
		return "INT UNSIGNED"
	case dbcodec.KindUint64:
		return "BIGINT UNSIGNED"

	case dbcodec.KindBoolean:
		return "TINYINT(1)"
	case dbcodec.KindFloat32:
		return "FLOAT"
	case dbcodec.KindFloat64:
		return "DOUBLE"
	case dbcodec.KindBinary:
		return "BLOB"
	case dbcodec.KindJSON:
		return "LONGTEXT"
	case dbcodec.KindDate:
		return "DATE"
	case dbcodec.KindDateTime:
		return "DATETIME"
	default:
		panic("unknown kind:" + kind)
	}
}

func (d MySQL) CreateTableIfNotExists(table string) string {
	return "CREATE TABLE IF NOT EXISTS " + d.QuoteIdentifier(table)
}

//	func (d MySQL) addColumnIfNotExists(table string, col string) string {
//		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s", d.QuoteIdentifier(table), d.QuoteIdentifier(col))
//	}
func (d MySQL) ColumnString(fs *dbschema.ColumnSchema) string {
	var sb strings.Builder
	sb.WriteString(d.QuoteIdentifier(fs.Name))
	sb.WriteString(" ")
	baseType := d.ColumnType(fs.Kind, fs.Size)
	sb.WriteString(baseType)
	if fs.NotNull {
		sb.WriteString(" NOT NULL")
	}
	if fs.Unique {
		sb.WriteString(" UNIQUE")
	}
	if fs.IsPrimaryKey {
		sb.WriteString(" PRIMARY KEY")
	}
	if fs.AutoIncrement {
		sb.WriteString(" AUTO_INCREMENT")
	}
	if dv := fs.Default; dv != nil {
		sb.WriteString(" DEFAULT ")
		switch dv.Type {
		case dbschema.DefaultValueTypeNumber, dbschema.DefaultValueTypeFn:
			sb.WriteString(dv.Value)
		case dbschema.DefaultValueTypeString:
			sb.WriteString(d.QuoteIdentifier(fs.Default.Value))
		default:
			panic(fmt.Sprintf("unknown default value type: %q", dv.Type))
		}
	}
	return sb.String()
}

var _ MigrateDialect = MySQL{}

func (d MySQL) Migrate(ctx context.Context, db DBCore, schema dbschema.TableSchema) error {
	sqlStr := createTableSQL(schema, d, d)
	_, err := db.ExecContext(ctx, sqlStr)
	return err
}
