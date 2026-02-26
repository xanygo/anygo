//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"context"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var _ dbtype.Dialect = (*SQLServerDialect)(nil)

type SQLServerDialect struct{}

// Name 返回方言名称
func (SQLServerDialect) Name() string {
	return "sqlserver"
}

// BindVar 返回 SQL Server 的占位符：@p1, @p2, ...
func (SQLServerDialect) BindVar(i int) string {
	return fmt.Sprintf("@p%d", i)
}

// QuoteIdentifier 使用方括号引用标识符
func (SQLServerDialect) QuoteIdentifier(s string) string {
	safe := strings.ReplaceAll(s, "]", "]]")
	return fmt.Sprintf("[%s]", safe)
}

// QuoteQualifiedIdentifier 支持 schema.table
func (d SQLServerDialect) QuoteQualifiedIdentifier(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = d.QuoteIdentifier(p)
	}
	return strings.Join(quoted, ".")
}

// LimitOffsetClause 生成 OFFSET/FETCH 子句
func (SQLServerDialect) LimitOffsetClause(limit, offset int) string {
	if limit < 0 && offset <= 0 {
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
func (d SQLServerDialect) PlaceholderList(n, start int) string {
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
func (SQLServerDialect) SupportReturning() bool {
	return false
}

func (SQLServerDialect) SupportLastInsertId() bool {
	return false
}

// ReturningClause SQL Server 用 OUTPUT 子句实现
func (SQLServerDialect) ReturningClause(columns ...string) string {
	if len(columns) == 0 {
		return "OUTPUT inserted.*"
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = "inserted." + fmt.Sprintf("[%s]", c)
	}
	return "OUTPUT " + strings.Join(quoted, ", ")
}

var _ dbtype.UpsertDialect = SQLServerDialect{}

// UpsertSQL 生成 SQL Server MERGE UPSERT
// table: 表名
// count: 数据条数
// cols: 所有字段
// conflictCols: 主键或唯一键字段
// updateCols: 需要更新的字段
// args: 对应参数值
// 返回：SQL string + 参数切片
func (d SQLServerDialect) UpsertSQL(table string, count int, cols, conflictCols, updateCols []string, returningCols []string) string {
	valPlaceholders := make([]string, len(cols))
	for c := range count {
		tmp := make([]string, len(cols))
		for i := range cols {
			tmp[i] = fmt.Sprintf("@p%d", c*i+1)
		}
		str := "(" + strings.Join(tmp, ",") + ")"
		valPlaceholders = append(valPlaceholders, str)
	}

	// ON 条件
	onCond := make([]string, len(conflictCols))

	placeholders := make([]string, len(cols))
	for i, c := range conflictCols {
		c = d.QuoteIdentifier(c)
		onCond[i] = fmt.Sprintf("target.[%s] = source.[%s]", c, c)
		placeholders[i] = fmt.Sprintf("source.%s", c)
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
USING (VALUES %s) AS source (%s)
ON %s
WHEN MATCHED THEN UPDATE SET %s
WHEN NOT MATCHED THEN INSERT (%s) VALUES (%s)
%s;`,
		table,
		strings.Join(valPlaceholders, ","), // VALUES 占位
		strings.Join(cols, ", "),           // source 列
		strings.Join(onCond, " AND "),      // ON 条件
		strings.Join(assigns, ", "),        // UPDATE
		strings.Join(cols, ", "),           // INSERT 列
		strings.Join(placeholders, ", "),   // INSERT VALUES
		output,                             // OUTPUT
	)

	return sqlStr
}

var _ dbtype.SchemaDialect = SQLServerDialect{}

// ColumnType 映射通用类型到 SQL Server 类型
func (SQLServerDialect) ColumnType(kind dbtype.Kind, size int) string {
	switch kind {
	case dbtype.KindString:
		if size <= 0 {
			return "TEXT"
		}
		return fmt.Sprintf("VARCHAR(%d)", size)

	case dbtype.KindUint8:
		return "TINYINT"
	case dbtype.KindUint16, dbtype.KindInt32:
		return "INT"
	case dbtype.KindInt, dbtype.KindUint, dbtype.KindUint32, dbtype.KindInt64:
		return "BIGINT"
	case dbtype.KindInt8, dbtype.KindInt16:
		return "SMALLINT"
	case dbtype.KindUint64:
		return "NUMERIC(20,0)"

	case dbtype.KindBoolean:
		return "BIT"

	case dbtype.KindFloat32:
		return "REAL"
	case dbtype.KindFloat64:
		return "FLOAT"

	case dbtype.KindBinary:
		return "VARBINARY(MAX)"

	case dbtype.KindJSON:
		return "NVARCHAR(MAX)" // SQL Server 2016+ 可以用 JSON 函数处理
	default:
		panic("unknown kind:" + kind)
	}
}

func (d SQLServerDialect) CreateTableIfNotExists(table string) string {
	return "IF NOT EXISTS (SELECT * FROM sysobjects where id = object_id('" +
		table + "') and OBJECTPROPERTY(id, 'IsUserTable') = 1 ) CREATE TABLE " + d.QuoteIdentifier(table)
}

func (d SQLServerDialect) ColumnString(fs dbtype.ColumnSchema) string {
	var sb strings.Builder
	sb.WriteString(d.QuoteIdentifier(fs.Name))
	sb.WriteString(" ")
	baseType := fs.Native
	if baseType == "" {
		baseType = d.ColumnType(fs.Kind, fs.Size)
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

var _ dbtype.MigrateDialect = SQLServerDialect{}

func (d SQLServerDialect) Migrate(ctx context.Context, db dbtype.DBCore, schema dbtype.TableSchema) error {
	sqlStr := createTableSQL(schema, d, d)
	_, err := db.ExecContext(ctx, sqlStr)
	return err
}
