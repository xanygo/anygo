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
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// Postgres 实现 Dialect 接口
type Postgres struct{}

// Name 返回方言名称
func (Postgres) Name() string {
	return "postgres"
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

	valPlaceholders := make([]string, len(cols))
	for c := 0; c < count; c++ {
		tmp := make([]string, len(cols))
		for i := range cols {
			tmp[i] = fmt.Sprintf("$%d", c*i+1)
		}
		str := "(" + strings.Join(tmp, ",") + ")"
		valPlaceholders = append(valPlaceholders, str)
	}

	updateAssignments := make([]string, len(updateCols))
	for i, c := range updateCols {
		c = d.QuoteIdentifier(c)
		updateAssignments[i] = fmt.Sprintf("%s = EXCLUDED.%s", c, c)
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s",
		table,
		colList,
		strings.Join(valPlaceholders, ","),
		strings.Join(conflictCols, ", "),
		strings.Join(updateAssignments, ", "),
	)

	if len(returningCols) > 0 {
		sqlStr += " RETURNING " + strings.Join(returningCols, ", ")
	}
	return sqlStr
}

var _ dbtype.SchemaDialect = Postgres{}

// ColumnType 映射通用类型到 Postgres 类型
func (Postgres) ColumnType(kind dbtype.Kind, size int) string {
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
		baseType = d.ColumnType(fs.Kind, fs.Size)
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
		case dbtype.DefaultValueTypeNumber, dbtype.DefaultValueTypeFn:
			sb.WriteString(dv.Value)
		case dbtype.DefaultValueTypeString:
			sb.WriteString(d.QuoteIdentifier(fs.Default.Value))
		default:
			panic(fmt.Sprintf("unknown default value type: %v", dv.Type))
		}
	}
	return sb.String()
}

var _ dbtype.MigrateDialect = Postgres{}

func (d Postgres) Migrate(ctx context.Context, db dbtype.DBCore, schema dbtype.TableSchema) error {
	sqlStr := createTableSQL(schema, d, d)
	_, err := db.ExecContext(ctx, sqlStr)
	return err
}

var _ dbtype.CoderDialect = Postgres{}

func (d Postgres) ColumnCodec(p reflect.Type) (dbtype.Codec, error) {
	switch p.Kind() {
	default:
		return nil, nil
	case reflect.Slice, reflect.Array:
		if zreflect.IsBasicKind(p.Elem().Kind()) {
			return pgAnyArrayCodec{}, nil
		}
		return nil, nil
	}
}

var _ dbtype.Codec = pgAnyArrayCodec{}

// pgAnyArrayCodec 数组类型的编解码功能
// 当字段这样定义包含 type:auto_json 的时候，会使用：
//
//	Scores []int `db:"scores,type:auto_json,native:int[]"`
type pgAnyArrayCodec struct{}

func (p pgAnyArrayCodec) Name() string {
	return "pgx_array"
}

func (p pgAnyArrayCodec) Encode(a any) (any, error) {
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
