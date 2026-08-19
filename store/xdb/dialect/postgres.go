//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// Postgres 实现 Dialect 接口
type Postgres struct{}

// Name 返回方言名称
func (Postgres) Name() string {
	return "postgres"
}

func (Postgres) RandomOrder() string {
	return "RANDOM()"
}

// BindVar 返回 Postgres 的占位符 `$1`, `$2`, ...
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

func (Postgres) LimitOffsetRequiresOrderBy() bool {
	return false
}

// LimitOffsetClause 生成 LIMIT/OFFSET 子句
func (Postgres) LimitOffsetClause(limit, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case limit == 0 && offset == 0:
		return ""
	case offset > 0:
		// PostgreSQL 不允许 LIMIT ALL OFFSET ?，需明确写 ALL
		return fmt.Sprintf("LIMIT ALL OFFSET %d", offset)
	default:
		return ""
	}
}

// PlaceholderList 返回 n 个占位符 ($1, $2, ...)
// start 应从 1 开始计数
func (Postgres) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	holders := make([]string, n)
	for i := range n {
		holders[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(holders, ", ")
}

// SupportReturning 返回 true
func (Postgres) SupportReturning() bool {
	return true
}

func (Postgres) SupportLastInsertId() bool {
	// LastInsertId is not supported by this driver
	return false
}

// ReturningClause 生成 RETURNING 子句
func (d Postgres) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "RETURNING *"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = d.QuoteIdentifier(c)
	}
	return "RETURNING " + strings.Join(quoted, ", ")
}

var _ dbtype.UpsertDialect = (*Postgres)(nil)

func (d Postgres) UpsertSQL(table string, count int, cols, conflictCols, updateCols []string, returningCols []string) string {
	colList := strings.Join(xslice.MapFunc(cols, d.QuoteIdentifier), ",")

	valPlaceholders := make([]string, 0, len(cols))
	for c := range count {
		str := "(" + d.PlaceholderList(len(cols), c*len(cols)+1) + ")"
		valPlaceholders = append(valPlaceholders, str)
	}

	updateAssignments := make([]string, len(updateCols))
	for i, c := range updateCols {
		c = d.QuoteIdentifier(c)
		updateAssignments[i] = fmt.Sprintf("%s = EXCLUDED.%s", c, c)
	}

	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s)",
		d.QuoteIdentifier(table),
		colList,
		strings.Join(valPlaceholders, ","),
		strings.Join(xslice.MapFunc(conflictCols, d.QuoteIdentifier), ","),
	)

	if len(updateAssignments) > 0 {
		sqlStr += " DO UPDATE SET " + strings.Join(updateAssignments, ", ")
	} else {
		sqlStr += " DO NOTHING "
	}

	if len(returningCols) > 0 {
		sqlStr += " RETURNING " + strings.Join(returningCols, ", ")
	}
	return sqlStr
}

var _ dbtype.SchemaDialect = Postgres{}

// ColumnKindType 映射通用类型到 Postgres 类型
func (Postgres) ColumnKindType(kind dbtype.Kind, size int) string {
	switch kind {
	case dbtype.KindString:
		if size <= 0 {
			return "TEXT"
		}
		return fmt.Sprintf("VARCHAR(%d)", size)
	case dbtype.KindInt8, dbtype.KindInt16, dbtype.KindUint8:
		return "SMALLINT"
	case dbtype.KindInt, dbtype.KindInt32, dbtype.KindUint16:
		return "INTEGER"
	case dbtype.KindUint, dbtype.KindInt64, dbtype.KindUint32:
		return "BIGINT"
	case dbtype.KindUint64:
		return "BIGINT"
		// return "NUMERIC(20,0)" // NUMERIC 不支持自增长
	case dbtype.KindBinary:
		return "BYTEA"
	case dbtype.KindBoolean:
		return "BOOLEAN"
	case dbtype.KindFloat32:
		return "REAL"
	case dbtype.KindFloat64:
		return "DOUBLE PRECISION"
	case dbtype.KindJSON:
		return "JSONB"
	case dbtype.KindDate:
		return "DATE"
	case dbtype.KindDateTime:
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func (d Postgres) EncodeValue(value any) (any, error) {
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
	default:
		return rv.Interface(), nil
	}
}

var _ dbtype.CoderDialect = Postgres{}

var arrCodec = pgAnyArrayCodec{}

func (Postgres) ColumnCodec(ft reflect.Type) (dbtype.Kind, dbtype.Codec, string) {
	switch ft.Kind() {
	case reflect.Slice, reflect.Array:
		switch ft.Elem().Kind() {
		case reflect.String:
			return dbtype.KindArray, arrCodec, "text[]"
		case reflect.Uint8: // 实际就是 []byte
			return dbtype.KindBinary, dbcodec.Binary{}, "BYTEA"
		case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint32:
			return dbtype.KindArray, arrCodec, "bigint[]"
		case reflect.Int8, reflect.Int16:
			return dbtype.KindArray, arrCodec, "smallint[]"
		case reflect.Uint16, reflect.Int32:
			return dbtype.KindArray, arrCodec, "integer[]"
		case reflect.Uint64:
			return dbtype.KindArray, arrCodec, "NUMERIC(20,0)[]"
		case reflect.Float32:
			return dbtype.KindArray, arrCodec, "REAL[]"
		case reflect.Float64:
			return dbtype.KindArray, arrCodec, "DOUBLE PRECISION[]"
		default:
			// pass
		}
	default:
		// pass
	}
	return dbtype.KindInvalid, nil, ""
}

func (d Postgres) CreateTableIfNotExists(table string) string {
	return "CREATE TABLE IF NOT EXISTS " + d.QuoteIdentifier(table)
}

// func (d Postgres) addColumnIfNotExists(table string, col string) string {
//	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s", d.QuoteIdentifier(table), d.QuoteIdentifier(col))
// }

// autoIncrementColumnType Postgres 自增列类型
func (Postgres) autoIncrementColumnType(baseType string) string {
	switch baseType {
	case "INTEGER":
		return "SERIAL" // 32 位自增
	case "SMALLINT":
		return "SMALLSERIAL"
	case "BIGINT":
		return "BIGSERIAL" // 64 位自增
	default:
		return baseType
	}
}

func (d Postgres) ColumnString(fs dbtype.ColumnSchema) string {
	var sb strings.Builder
	sb.WriteString(d.QuoteIdentifier(fs.Name))
	sb.WriteString(" ")
	baseType := fs.Native
	if baseType == "" {
		baseType = d.ColumnKindType(fs.Kind, fs.Size)
	}
	if fs.AutoIncrement {
		sb.WriteString(d.autoIncrementColumnType(baseType))
	} else {
		sb.WriteString(baseType)
	}
	if fs.NotNull {
		sb.WriteString(" NOT NULL")
	}
	if fs.Unique {
		sb.WriteString(" UNIQUE")
	}
	if fs.IsPrimaryKey {
		sb.WriteString(" PRIMARY KEY")
	}
	if dv := fs.Default; dv != nil {
		sb.WriteString(" DEFAULT ")
		switch dv.Type {
		case dbtype.DefaultValueTypeNumber:
			sb.WriteString(dv.Value)
		case dbtype.DefaultValueTypeFn:
			// pgx 内置支持 CURRENT_DATE (2026-08-08),CURRENT_TIMESTAMP (2026-08-08 08:08:08)
			sb.WriteString(dv.Value)
		case dbtype.DefaultValueTypeString:
			sb.WriteString(d.QuoteIdentifier(fs.Default.Value))
		default:
			panic(fmt.Sprintf("unknown default value type: %v", dv.Type))
		}
	} else if fs.NotNull && !fs.AutoIncrement {
		if baseType == "JSONB" || strings.HasSuffix(baseType, "]") {
			sb.WriteString(" DEFAULT '{}'")
		} else if strings.HasPrefix(baseType, "VARCHAR") {
			sb.WriteString(" DEFAULT ''")
		} else {
			switch baseType {
			case "BYTEA", "TEXT":
				sb.WriteString(" DEFAULT ''")
			case "BIGINT", "INTEGER", "SMALLINT", "REAL", "DOUBLE PRECISION":
				sb.WriteString(" DEFAULT 0")
			case "BOOLEAN":
				sb.WriteString(" DEFAULT false")
			}
		}
	}
	return sb.String()
}

func (d Postgres) UniqIndex(name string, columns []string) string {
	return fmt.Sprintf("CONSTRAINT %s UNIQUE(%s)", d.QuoteIdentifier(name), quoteIdentifiersJoin(d, columns))
}

func (d Postgres) AlterCreateIndex(indexType string, name string, table string, columns []string) string {
	name += "_" + table // 避免不同表的索引名称重复
	return fmt.Sprintf("CREATE %s IF NOT EXISTS %s on %s(%s)",
		indexType, d.QuoteIdentifier(name), d.QuoteIdentifier(table), quoteIdentifiersJoin(d, columns))
}

var _ dbtype.MigrateDialect = Postgres{}

func (d Postgres) Migrate(ctx context.Context, db dbtype.DBCore, schema dbtype.TableSchema) error {
	sqls := createTableSQLList(schema, d, d)
	for _, sql := range sqls {
		_, err := db.ExecContext(ctx, sql)
		if err != nil {
			return fmt.Errorf("postgres migrate SQL %q: %w", sql, err)
		}
	}
	return nil
}

var _ dbtype.DescDialect = Postgres{}

func (d Postgres) CurrentDatabase(ctx context.Context, q dbtype.Queryer) (string, error) {
	const str = "SELECT current_database()"
	return queryOneString(ctx, q, str)
}

func (d Postgres) Databases(ctx context.Context, q dbtype.Queryer) ([]string, error) {
	const str = `SELECT datname FROM pg_database WHERE datistemplate = false`
	return querySliceString(ctx, q, str)
}

func (d Postgres) Tables(ctx context.Context, q dbtype.Queryer) ([]string, error) {
	const str = `SELECT table_name FROM information_schema.tables
WHERE table_schema = current_schema() AND table_type = 'BASE TABLE' ORDER BY table_name;`
	return querySliceString(ctx, q, str)
}

func (d Postgres) TableExists(ctx context.Context, q dbtype.Queryer, table string) (bool, error) {
	const str = `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1 )`
	return queryBool(ctx, q, str, table)
}

func (d Postgres) TableColumns(ctx context.Context, q dbtype.Queryer, table string) ([]string, error) {
	const str = `SELECT column_name FROM information_schema.columns
WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position`
	return querySliceString(ctx, q, str, table)
}

var _ dbtype.Codec = pgAnyArrayCodec{}

// pgAnyArrayCodec 数组类型的编解码功能
//
//	Scores []int `db:"scores,codec:auto_json"`
type pgAnyArrayCodec struct{}

func (p pgAnyArrayCodec) Name() string {
	return "pgx_array"
}

func (p pgAnyArrayCodec) Encode(a any) (any, error) {
	if a == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(a)
	if rv.Kind() == reflect.Array {
		return zreflect.ArrayToSlice(rv), nil
	}
	return a, nil
}

func (p pgAnyArrayCodec) Decode(b string, a any) error {
	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be pointer to slice, got %T", a)
	}
	str := p.arrayToJSONSimple(b)
	return json.Unmarshal([]byte(str), a)
}

func (p pgAnyArrayCodec) Kind() dbtype.Kind {
	return dbtype.KindArray
}

func (p pgAnyArrayCodec) arrayToJSONSimple(s string) string {
	var buf strings.Builder
	inString := false
	escape := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escape {
			buf.WriteByte(ch)
			escape = false
			continue
		}

		if ch == '\\' {
			escape = true
			buf.WriteByte(ch)
			continue
		}

		if ch == '"' {
			inString = !inString
			buf.WriteByte(ch)
			continue
		}

		if !inString {
			switch ch {
			case '{':
				buf.WriteByte('[')
				continue
			case '}':
				buf.WriteByte(']')
				continue
			}
		}

		buf.WriteByte(ch)
	}

	return buf.String()
}
