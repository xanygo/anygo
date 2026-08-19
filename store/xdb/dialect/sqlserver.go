//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var _ dbtype.Dialect = (*SQLServer)(nil)

type SQLServer struct{}

// Name 返回方言名称
func (SQLServer) Name() string {
	return "sqlserver"
}

func (SQLServer) RandomOrder() string {
	return "NEWID()"
}

// BindVar 返回 SQL Server 的占位符：@p1, @p2, ...
func (SQLServer) BindVar(i int) string {
	return fmt.Sprintf("@p%d", i)
}

// QuoteIdentifier 使用方括号引用标识符
func (SQLServer) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, "]", "]]")
	return fmt.Sprintf("[%s]", safe)
}

// QuoteQualifiedIdentifier 支持 schema.table
func (d SQLServer) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

func (SQLServer) LimitOffsetRequiresOrderBy() bool {
	// SELECT * FROM [user] ORDER BY id OFFSET 20 ROWS;
	// SQLServer 必须要有 order by 语句
	return true
}

// LimitOffsetClause 生成 OFFSET/FETCH 子句
func (SQLServer) LimitOffsetClause(limit, offset int) string {
	if limit <= 0 && offset <= 0 {
		return ""
	}
	if limit < 0 {
		limit = 2147483647 // SQL Server 最大 INT
	}
	if offset < 0 {
		offset = 0
	}
	return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
}

// PlaceholderList 返回占位符列表
func (d SQLServer) PlaceholderList(n, start int) string {
	if n <= 0 {
		return ""
	}
	holders := make([]string, n)
	for i := range n {
		holders[i] = d.BindVar(start + i)
	}
	return strings.Join(holders, ", ")
}

// SupportReturning SQL Server 不直接支持 RETURNING
func (SQLServer) SupportReturning() bool {
	return false
}

func (SQLServer) SupportLastInsertId() bool {
	return false
}

// ReturningClause SQL Server 用 OUTPUT 子句实现
func (d SQLServer) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "OUTPUT inserted.*"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = "inserted." + d.QuoteIdentifier(c)
	}
	return "OUTPUT " + strings.Join(quoted, ", ")
}

var _ dbtype.UpsertDialect = SQLServer{}

// UpsertSQL 生成 SQL Server MERGE UPSERT
// table: 表名
// count: 数据条数
// cols: 所有字段
// conflictCols: 主键或唯一键字段
// updateCols: 需要更新的字段
// args: 对应参数值
// 返回：SQL string + 参数切片
func (d SQLServer) UpsertSQL(table string, count int, cols, conflictCols, updateCols []string, returningCols []string) string {
	valPlaceholders := make([]string, 0, count)
	for c := range count {
		str := "(" + d.PlaceholderList(len(cols), c*len(cols)+1) + ")"
		valPlaceholders = append(valPlaceholders, str)
	}

	// ON 条件
	onCond := make([]string, len(conflictCols))

	for i, c := range conflictCols {
		c = d.QuoteIdentifier(c)
		onCond[i] = fmt.Sprintf("target.%s = source.%s", c, c)
	}

	// UPDATE 赋值
	assigns := make([]string, len(updateCols))
	for i, c := range updateCols {
		assigns[i] = fmt.Sprintf("target.[%s] = source.[%s]", c, c)
	}

	// OUTPUT 子句
	var output string
	if len(returningCols) > 0 {
		tmp := make([]string, len(returningCols))
		for i, c := range returningCols {
			tmp[i] = fmt.Sprintf("inserted.[%s]", c)
		}
		output = "OUTPUT " + strings.Join(tmp, ", ")
	}

	// MERGE INTO users AS t
	// USING (VALUES
	//    (1, 'Tom', 10),
	//    (2, 'Bob', 15),
	//    (3, 'Amy', 20)
	// ) AS s(id, name, score)
	//    ON t.id = s.id
	// WHEN MATCHED THEN
	//    UPDATE SET
	//        t.name = s.name,
	//        t.score = s.score
	// WHEN NOT MATCHED THEN
	//    INSERT (id, name, score)
	//    VALUES (s.id, s.name, s.score);

	// 生成完整 MERGE SQL
	sqlStr := fmt.Sprintf(
		`MERGE INTO %s AS target
USING (VALUES %s) 
AS source (%s)
ON %s`,
		d.QuoteIdentifier(table),
		strings.Join(valPlaceholders, ","), // VALUES 占位
		strings.Join(cols, ", "),           // source 列
		strings.Join(onCond, " AND "),      // ON 条件
	)
	if len(assigns) > 0 {
		sqlStr += "\n WHEN MATCHED THEN UPDATE SET " + strings.Join(assigns, ", ") // UPDATE
	}

	placeholders := make([]string, len(cols))
	for i, c := range cols {
		placeholders[i] = fmt.Sprintf("source.%s", d.QuoteIdentifier(c))
	}

	sqlStr += "\n"
	sqlStr += fmt.Sprintf(`WHEN NOT MATCHED THEN INSERT (%s) VALUES (%s)`,
		strings.Join(cols, ", "),         // INSERT 列
		strings.Join(placeholders, ", "), // INSERT VALUES
	)
	sqlStr += output // OUTPUT
	sqlStr += ";"

	return sqlStr
}

var _ dbtype.SchemaDialect = SQLServer{}

// ColumnKindType 映射通用类型到 SQL Server 类型
func (SQLServer) ColumnKindType(kind dbtype.Kind, size int) string {
	switch kind {
	case dbtype.KindString:
		if size <= 0 {
			return "NVARCHAR(MAX)"
		}
		return fmt.Sprintf("NVARCHAR(%d)", size)

	case dbtype.KindUint8:
		return "TINYINT"
	case dbtype.KindUint16, dbtype.KindInt32:
		return "INT"
	case dbtype.KindInt, dbtype.KindUint, dbtype.KindUint32, dbtype.KindInt64:
		return "BIGINT"
	case dbtype.KindInt8, dbtype.KindInt16:
		return "SMALLINT"
	case dbtype.KindUint64:
		return "BIGINT"
		// return "NUMERIC(20,0)"

	case dbtype.KindBoolean:
		return "BIT"

	case dbtype.KindFloat32:
		return "REAL"
	case dbtype.KindFloat64:
		return "FLOAT"

	case dbtype.KindBinary:
		if size > 0 {
			return fmt.Sprintf("BINARY(%d)", size)
		}
		return "VARBINARY(MAX)"

	case dbtype.KindJSON:
		return "NVARCHAR(MAX)" // SQL Server 2016+ 可以用 JSON 函数处理
	case dbtype.KindDateTime:
		return "DATETIME2" // 可存储  0001-9999
	case dbtype.KindDate:
		return "DATE"
	default:
		panic("unknown kind:" + kind)
	}
}

func (d SQLServer) CreateTableIfNotExists(table string) string {
	return "IF NOT EXISTS (SELECT * FROM sysobjects where id = object_id('" +
		table + "') and OBJECTPROPERTY(id, 'IsUserTable') = 1 ) CREATE TABLE " + d.QuoteIdentifier(table)
}

func (d SQLServer) ColumnString(fs dbtype.ColumnSchema) string {
	var sb strings.Builder
	sb.WriteString(d.QuoteIdentifier(fs.Name))
	sb.WriteString(" ")
	baseType := fs.Native
	if baseType == "" {
		baseType = d.ColumnKindType(fs.Kind, fs.Size)
	}
	sb.WriteString(baseType)
	if fs.AutoIncrement && strings.Contains(baseType, "INT") {
		sb.WriteString(" IDENTITY(1,1)")
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
			switch dv.Value {
			case dbtype.CurrentDate: // 2026-08-08
				sb.WriteString("CAST(GETDATE() AS date)")
			case dbtype.CurrentTimestamp: // 2026-08-08 08:08:08
				sb.WriteString("SYSDATETIME()") // 用 datetime2 类型存储
			default:
				sb.WriteString(dv.Value)
			}
		case dbtype.DefaultValueTypeString:
			sb.WriteString(d.QuoteIdentifier(fs.Default.Value))
		default:
			panic(fmt.Sprintf("unknown default value type: %v", dv.Type))
		}
	} else if fs.NotNull && !fs.AutoIncrement {
		if strings.HasSuffix(baseType, "INT") || baseType == "REAL" || baseType == "FLOAT" {
			sb.WriteString(" DEFAULT 0")
		} else if strings.HasPrefix(baseType, "NVARCHAR") || strings.HasPrefix(baseType, "VARBINARY") {
			sb.WriteString(" DEFAULT ''")
		}
	}
	return sb.String()
}

func (d SQLServer) UniqIndex(name string, columns []string) string {
	return fmt.Sprintf("CONSTRAINT %s UNIQUE(%s)", d.QuoteIdentifier(name), quoteIdentifiersJoin(d, columns))
}

func (d SQLServer) AlterCreateIndex(indexType string, name string, table string, columns []string) string {
	// 不支持 IF NOT EXISTS
	name += "_" + table // 避免不同表的索引名称重复
	return fmt.Sprintf("CREATE %s %s on %s(%s)",
		indexType, d.QuoteIdentifier(name), d.QuoteIdentifier(table), quoteIdentifiersJoin(d, columns))
}

var _ dbtype.MigrateDialect = SQLServer{}

func (d SQLServer) Migrate(ctx context.Context, db dbtype.DBCore, schema dbtype.TableSchema) error {
	sqlStr := createTableSQL(schema, d, d)
	_, err := db.ExecContext(ctx, sqlStr)
	if err != nil {
		return fmt.Errorf("sqlserver Migrate SQL %q: %w", sqlStr, err)
	}
	return nil
}

var _ dbtype.DescDialect = SQLServer{}

func (d SQLServer) CurrentDatabase(ctx context.Context, q dbtype.Queryer) (string, error) {
	const str = "SELECT DB_NAME()"
	return queryOneString(ctx, q, str)
}

func (d SQLServer) Databases(ctx context.Context, q dbtype.Queryer) ([]string, error) {
	const str = `SELECT name FROM sys.databases`
	return querySliceString(ctx, q, str)
}

func (d SQLServer) Tables(ctx context.Context, q dbtype.Queryer) ([]string, error) {
	const str = `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE'`
	return querySliceString(ctx, q, str)
}

func (d SQLServer) TableExists(ctx context.Context, q dbtype.Queryer, table string) (bool, error) {
	const str = `SELECT CASE WHEN EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = @p1 ) THEN 1 ELSE 0 END`
	return queryBool(ctx, q, str, table)
}

func (d SQLServer) TableColumns(ctx context.Context, q dbtype.Queryer, table string) ([]string, error) {
	const str = `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = @p1 ORDER BY ORDINAL_POSITION`
	return querySliceString(ctx, q, str, table)
}

func (d SQLServer) EncodeValue(value any) (any, error) {
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
