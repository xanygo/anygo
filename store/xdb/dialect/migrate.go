//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-15

package dialect

import (
	"sort"
	"strings"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

func createTableSQL(ts dbtype.TableSchema, d dbtype.Dialect, sd dbtype.SchemaDialect) string {
	pks := ts.PKColumns()

	str := sd.CreateTableIfNotExists(ts.Table) + " (\n"

	var lines []string
	indexMap := map[string][]*dbtype.IndexSchema{}
	uniqIndexMap := map[string][]*dbtype.IndexSchema{}
	for _, field := range ts.Columns {
		if len(pks) > 1 && field.IsPrimaryKey {
			field.IsPrimaryKey = false // 后面统一生成联合主键
		}
		tmp := sd.ColumnString(field)
		lines = append(lines, tmp)
		for _, index := range field.Indexes {
			indexName := index.IndexName
			if index.Unique {
				uniqIndexMap[indexName] = append(uniqIndexMap[indexName], index)
			} else {
				indexMap[indexName] = append(indexMap[indexName], index)
			}
		}
	}

	for indexName, indexes := range indexMap {
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].FieldOrder < indexes[j].FieldOrder
		})
		var names []string
		for _, index := range indexes {
			names = append(names, d.QuoteIdentifier(index.FieldName))
		}
		tmp := "Index " + indexName + "(" + strings.Join(names, ",") + ")"
		lines = append(lines, tmp)
	}

	for indexName, indexes := range uniqIndexMap {
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].FieldOrder < indexes[j].FieldOrder
		})
		var names []string
		for _, index := range indexes {
			names = append(names, index.FieldName)
		}
		tmp := sd.UniqIndex(indexName, names)
		lines = append(lines, tmp)
	}
	if len(pks) > 1 {
		var pkCols []string
		for _, field := range pks {
			pkCols = append(pkCols, d.QuoteIdentifier(field.Name))
		}
		tmp := "PRIMARY KEY(" + strings.Join(pkCols, ",") + ")"
		lines = append(lines, tmp)
	}
	str += strings.Join(lines, ",\n") + ")"
	return str
}

// 返回创建表和创建索引的语句是独立的。sqlite 使用中
func createTableSQLList(ts dbtype.TableSchema, d dbtype.Dialect, sd dbtype.SchemaDialect) []string {
	var result []string
	pks := ts.PKColumns()
	indexMap := map[string][]*dbtype.IndexSchema{}
	uniqIndexMap := map[string][]*dbtype.IndexSchema{}

	// 生成基础的创建 Table 的 schema
	{
		str := sd.CreateTableIfNotExists(ts.Table) + " (\n"
		var lines []string
		for _, field := range ts.Columns {
			if len(pks) > 1 && field.IsPrimaryKey {
				field.IsPrimaryKey = false // 后面统一生成联合主键
			}
			tmp := sd.ColumnString(field)
			lines = append(lines, tmp)
			for _, index := range field.Indexes {
				indexName := index.IndexName
				if index.Unique {
					uniqIndexMap[indexName] = append(uniqIndexMap[indexName], index)
				} else {
					indexMap[indexName] = append(indexMap[indexName], index)
				}
			}
		}

		if len(pks) > 1 {
			var pkCols []string
			for _, field := range pks {
				pkCols = append(pkCols, d.QuoteIdentifier(field.Name))
			}
			tmp := "PRIMARY KEY(" + strings.Join(pkCols, ",") + ")"
			lines = append(lines, tmp)
		}

		str += strings.Join(lines, ",\n") + ")"
		result = append(result, str)
	}

	for indexName, indexes := range indexMap {
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].FieldOrder < indexes[j].FieldOrder
		})
		var cols []string
		for _, index := range indexes {
			cols = append(cols, index.FieldName)
		}
		str := sd.AlterCreateIndex("INDEX", indexName, ts.Table, cols)
		result = append(result, str)
	}

	for indexName, indexes := range uniqIndexMap {
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].FieldOrder < indexes[j].FieldOrder
		})
		var cols []string
		for _, index := range indexes {
			cols = append(cols, index.FieldName)
		}
		str := sd.AlterCreateIndex("UNIQUE INDEX", indexName, ts.Table, cols)
		result = append(result, str)
	}

	return result
}

func quoteIdentifiersJoin(d dbtype.Dialect, cols []string) string {
	return strings.Join(xslice.MapFunc(cols, d.QuoteIdentifier), ",")
}
