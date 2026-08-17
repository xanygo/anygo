//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbtype

import (
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/xanygo/anygo/xerror"
)

// ErrNoPK 错误：没有主键
var ErrNoPK = errors.New("no primary key column")

type TableSchema struct {
	Table       string                  // 数据库表名，可能为空
	Columns     []ColumnSchema          // 字段列表
	Name2Column map[string]ColumnSchema // 数据库字段名 <---> 字段属性的映射
	ColumnNames []string                // 数据库中的字段名
	TagName     string
}

func (ts *TableSchema) ColumnByName(name string) (ColumnSchema, error) {
	f, ok := ts.Name2Column[name]
	if ok {
		return f, nil
	}
	return f, fmt.Errorf("column %q %w", name, xerror.NotFound)
}

// PKColumns 返回表中的主键字段，允许有多个字段（多个字段联合为主键）。
func (ts *TableSchema) PKColumns() ColumnSchemas {
	var result []ColumnSchema
	for _, col := range ts.Columns {
		if col.IsPrimaryKey {
			result = append(result, col)
		}
	}
	return result
}

func (ts *TableSchema) FilterByGroup(group string) ColumnSchemas {
	if group == "" {
		return nil
	}
	var result ColumnSchemas
	for _, col := range ts.Columns {
		if col.IsGroup(group) {
			result = append(result, col)
		}
	}

	if len(result) == 0 {
		for _, col := range ts.Columns {
			if col.UniqueIndex != nil && col.UniqueIndex.IndexName == group {
				result = append(result, col)
			}
		}
	}
	return result
}

type ColumnSchemas []ColumnSchema

func (cs ColumnSchemas) Len() int {
	return len(cs)
}

func (cs ColumnSchemas) Names() []string {
	result := make([]string, 0, len(cs))
	for _, c := range cs {
		result = append(result, c.Name)
	}
	return result
}

type ColumnSchema struct {
	Name          string              // 数据库字段名
	IsPrimaryKey  bool                // 是否主键
	AutoIncrement bool                // 自增长
	Kind          Kind                // 数据类型
	Unique        bool                // 是否唯一键
	Index         *IndexSchema        // 索引的名称
	UniqueIndex   *IndexSchema        // 唯一索引
	Size          int                 // 定义列数据类型的大小或长度
	NotNull       bool                // 是否申明 not null 属性
	Codec         Codec               // 字段编解码器
	Native        string              // 数据库原生类型
	Default       *DefaultValueSchema // 默认值
	Auto          string              // 编码数据是自动化处理规则，可选值如，created，updated

	ReflectType reflect.Type // struct 中字段的类型
	Group       []string     // 分组标签
}

func (scf ColumnSchema) IsGroup(group string) bool {
	return slices.Contains(scf.Group, group)
}

func (scf ColumnSchema) String() string {
	return fmt.Sprintf("%#v", scf)
}

type IndexSchema struct {
	FieldName  string // 数据库字段名
	IndexName  string // 索引名
	FieldOrder int    // 字段在索引中的顺序
}

type DefaultValueSchema struct {
	// Type 值类型，可选值：number，string，fn
	// 当为 number、fn 时：拼接到 schema 里去的时候，直接拼接，不需要使用 "" 转义
	Type DefaultValueType

	Value string // 值的字符串形式
}

type DefaultValueType int8

const (
	DefaultValueTypeString DefaultValueType = iota
	DefaultValueTypeNumber
	DefaultValueTypeFn
)
