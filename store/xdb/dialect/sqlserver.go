//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dialect

import (
	"fmt"
	"strings"
)

var _ Dialect = (*SQLServerDialect)(nil)

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
	for i := 0; i < n; i++ {
		holders[i] = d.BindVar(start + i)
	}
	return strings.Join(holders, ", ")
}

// SupportsReturning SQL Server 不直接支持 RETURNING
func (SQLServerDialect) SupportsReturning() bool {
	return false
}

// SupportsUpsert 支持 MERGE 语法实现 UPSERT
func (SQLServerDialect) SupportsUpsert() bool {
	return true
}

// DefaultValueExpr 默认值关键字
func (SQLServerDialect) DefaultValueExpr() string {
	return "DEFAULT"
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

var _ UpsertDialect = SQLServerDialect{}

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
	for c := 0; c < count; c++ {
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

// AutoIncrementColumnType 返回自增列定义
func (SQLServerDialect) AutoIncrementColumnType(baseType string) string {
	switch strings.ToLower(baseType) {
	case "int", "integer":
		return "INT IDENTITY(1,1)"
	case "bigint":
		return "BIGINT IDENTITY(1,1)"
	default:
		return baseType
	}
}

// ColumnType 映射通用类型到 SQL Server 类型
func (SQLServerDialect) ColumnType(kind string, size int) string {
	switch strings.ToLower(kind) {
	case "string", "varchar":
		if size <= 0 {
			size = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", size)
	case "text":
		return "TEXT"
	case "int", "integer":
		return "INT"
	case "bigint":
		return "BIGINT"
	case "bool", "boolean":
		return "BIT"
	case "float", "double", "real":
		return "FLOAT"
	case "json":
		return "NVARCHAR(MAX)" // SQL Server 2016+ 可以用 JSON 函数处理
	default:
		return kind
	}
}

func (SQLServerDialect) CreateTableIfNotExists() string {
	return "IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='[TABLE]' AND xtype='U') CREATE TABLE"
}
