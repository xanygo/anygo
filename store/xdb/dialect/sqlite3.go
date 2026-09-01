//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// SQLite3 实现 Dialect 接口
type SQLite3 struct{}

// Name 返回方言名称
func (SQLite3) Name() string {
	return "sqlite3"
}

func (SQLite3) RandomOrder() string {
	return "RANDOM()"
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

func (SQLite3) LimitOffsetRequiresOrderBy() bool {
	return false
}

// LimitOffsetClause 生成 LIMIT/OFFSET 子句。
// SQLite3 支持标准写法 "LIMIT ? OFFSET ?"
func (SQLite3) LimitOffsetClause(limit, offset int) string {
	if limit <= 0 && offset <= 0 {
		return ""
	}
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
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

// SupportReturning 从 SQLite3 3.35.0 (2021-03) 起支持 RETURNING。
func (SQLite3) SupportReturning() bool {
	return true
}

func (SQLite3) SupportLastInsertId() bool {
	return true
}

// ReturningClause 生成 RETURNING 子句。
// 空 columns 表示 RETURNING *。
func (d SQLite3) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "RETURNING *"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = d.QuoteIdentifier(c)
	}
	return "RETURNING " + strings.Join(quoted, ", ")
}

var _ dbtype.UpsertDialect = SQLite3{}

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
	sqlStr += fmt.Sprintf(" ON CONFLICT (%s) ", strings.Join(xslice.MapFunc(conflictCols, d.QuoteIdentifier), ", "))
	if len(updateAssignments) > 0 {
		sqlStr += " DO UPDATE SET " + strings.Join(updateAssignments, ", ")
	} else {
		sqlStr += " DO NOTHING "
	}

	if len(returningCols) > 0 {
		sqlStr += " RETURNING " + strings.Join(xslice.MapFunc(returningCols, d.QuoteIdentifier), ",")
	}
	return sqlStr
}

var _ dbtype.SchemaDialect = SQLite3{}

func (SQLite3) ColumnKindType(kind dbtype.Kind, size int) string {
	switch kind {
	case dbtype.KindString:
		return "TEXT"
	case dbtype.KindInt, dbtype.KindInt8, dbtype.KindInt16, dbtype.KindInt32, dbtype.KindInt64,
		dbtype.KindUint, dbtype.KindUint8, dbtype.KindUint16, dbtype.KindUint32, dbtype.KindUint64:
		return "INTEGER"
	case dbtype.KindBoolean:
		return "INTEGER"
	case dbtype.KindFloat32, dbtype.KindFloat64:
		return "REAL"
	case dbtype.KindJSON:
		// SQLite3 3.38+ 支持 JSON 函数，但底层仍 TEXT
		return "TEXT"
	case dbtype.KindBinary:
		return "BLOB"
	default:
		return "TEXT"
	}
}

func (d SQLite3) ColumnString(fs dbtype.ColumnSchema) string {
	var sb strings.Builder
	sb.WriteString(d.QuoteIdentifier(fs.Name))
	sb.WriteString(" ")
	baseType := fs.Native
	if baseType == "" {
		baseType = d.ColumnKindType(fs.Kind, fs.Size)
	}
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
		sb.WriteString(" AUTOINCREMENT")
	}
	if dv := fs.Default; dv != nil {
		sb.WriteString(" DEFAULT ")
		switch dv.Type {
		case dbtype.DefaultValueTypeNumber:
			sb.WriteString(dv.Value)
		case dbtype.DefaultValueTypeFn:
			// sqlite 内置支持  CURRENT_DATE (2026-08-08),CURRENT_TIMESTAMP (2026-08-08 08:08:08)
			sb.WriteString(dv.Value)
		case dbtype.DefaultValueTypeString:
			sb.WriteString(d.QuoteIdentifier(fs.Default.Value))
		default:
			panic(fmt.Sprintf("unknown default value type: %v", dv.Type))
		}
	} else if fs.NotNull && !fs.AutoIncrement {
		switch baseType {
		case "TEXT", "BLOB":
			sb.WriteString(" DEFAULT ''")
		case "INTEGER", "REAL":
			sb.WriteString(" DEFAULT 0")
		}
	}

	return sb.String()
}

func (d SQLite3) UniqIndex(name string, columns []string) string {
	return fmt.Sprintf("CONSTRAINT %s UNIQUE(%s)", d.QuoteIdentifier(name), quoteIdentifiersJoin(d, columns))
}

func (d SQLite3) CreateTableIfNotExists(table string) string {
	return "CREATE TABLE IF NOT EXISTS " + d.QuoteIdentifier(table)
}

func (d SQLite3) AlterCreateIndex(indexType string, name string, table string, columns []string) string {
	name += "_" + table // 避免不同表的索引名称重复
	return fmt.Sprintf("CREATE %s IF NOT EXISTS %s on %s(%s)",
		indexType, d.QuoteIdentifier(name), d.QuoteIdentifier(table), quoteIdentifiersJoin(d, columns))
}

var _ dbtype.MigrateDialect = SQLite3{}

func (d SQLite3) Migrate(ctx context.Context, db dbtype.DBCore, schema dbtype.TableSchema) error {
	return doMigrate(ctx, d, db, schema)
}

var _ dbtype.DescDialect = SQLite3{}

func (d SQLite3) CurrentDatabase(ctx context.Context, q dbtype.Queryer) (string, error) {
	return "main", nil
}

func (d SQLite3) Databases(ctx context.Context, q dbtype.Queryer) ([]string, error) {
	rows, err := q.QueryContext(ctx, `PRAGMA database_list`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var seq int
		var name string
		var file string
		if err = rows.Scan(&seq, &name, &file); err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	return result, nil
}

func (d SQLite3) Tables(ctx context.Context, q dbtype.Queryer) ([]string, error) {
	const str = `SELECT name FROM sqlite_master WHERE type='table'`
	return querySliceString(ctx, q, str)
}

func (d SQLite3) TableExists(ctx context.Context, q dbtype.Queryer, table string) (bool, error) {
	const str = `SELECT EXISTS ( SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)`
	return queryBool(ctx, q, str, table)
}

func (d SQLite3) TableColumns(ctx context.Context, q dbtype.Queryer, table string) ([]string, error) {
	str := fmt.Sprintf("PRAGMA table_info(%s)", d.QuoteIdentifier(table))
	rows, err := q.QueryContext(ctx, str)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []string

	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notnull int
			def     any
			pk      int
		)

		err = rows.Scan(&cid, &name, &typ, &notnull, &def, &pk)

		if err != nil {
			return nil, err
		}

		columns = append(columns, name)
	}
	return columns, nil
}

func (d SQLite3) EncodeValue(value any) (any, error) {
	rv := reflect.ValueOf(value)
	for rv.IsValid() && rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil, nil
	}
	switch rv.Kind() {
	case reflect.Array:
		return zreflect.ArrayToSlice(rv), nil
	case reflect.Bool:
		if rv.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return rv.Interface(), nil
	}
}

func (d SQLite3) SavepointSQL(name string) string {
	return "SAVEPOINT " + d.QuoteIdentifier(name)
}

func (d SQLite3) RollbackToSavepointSQL(name string) string {
	return "ROLLBACK TO SAVEPOINT " + d.QuoteIdentifier(name)
}

func (d SQLite3) ReleaseSavepointSQL(name string) string {
	return "RELEASE SAVEPOINT " + d.QuoteIdentifier(name)
}
