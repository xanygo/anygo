//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbtype

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/xerror"
)

type TableSchema struct {
	Table       string                  // 数据库表名，可能为空
	Columns     []ColumnSchema          // 字段列表
	Name2Column map[string]ColumnSchema // 数据库字段名 <---> 字段属性的映射
	ColumnNames []string                // 数据库中的字段名
}

func (ts *TableSchema) ColumnByName(name string) (ColumnSchema, error) {
	f, ok := ts.Name2Column[name]
	if ok {
		return f, nil
	}
	return f, fmt.Errorf("column %q %w", name, xerror.NotFound)
}

var errNoPK = errors.New("no primary key column")

// PKColumn 查找主键字段(单字段主键)
func (ts *TableSchema) PKColumn() (z ColumnSchema, err error) {
	for _, col := range ts.Columns {
		if col.IsPrimaryKey {
			return col, nil
		}
	}
	return z, errNoPK
}

type ColumnSchema struct {
	Name          string              // 字段名
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
	ReflectType   reflect.Type
}

func (scf *ColumnSchema) String() string {
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
