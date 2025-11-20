//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-15

package dialect

import (
	"sort"
	"strings"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

func createTableSQL(ts dbtype.TableSchema, d dbtype.Dialect, sd dbtype.SchemaDialect) string {
	str := sd.CreateTableIfNotExists(ts.Table) + " (\n"

	var lines []string
	indexMap := map[string][]*dbtype.IndexSchema{}
	uniqIndexMap := map[string][]*dbtype.IndexSchema{}
	for _, field := range ts.Columns {
		tmp := sd.ColumnString(field)
		lines = append(lines, tmp)
		if field.Index != nil {
			indexName := field.Index.IndexName
			indexMap[indexName] = append(indexMap[indexName], field.Index)
		}
		if field.UniqueIndex != nil {
			indexName := field.UniqueIndex.IndexName
			uniqIndexMap[indexName] = append(uniqIndexMap[indexName], field.UniqueIndex)
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
			names = append(names, d.QuoteIdentifier(index.FieldName))
		}
		tmp := "UNIQUE KEY " + indexName + "(" + strings.Join(names, ",") + ")"
		lines = append(lines, tmp)
	}
	str += strings.Join(lines, ",\n") + ")"
	return str
}
