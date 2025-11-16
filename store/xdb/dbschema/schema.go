//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-15

package dbschema

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zcache"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
)

type TableSchema struct {
	Table       string
	Columns     []*ColumnSchema
	name2Column map[string]*ColumnSchema
}

func (ts *TableSchema) ColumnByName(name string) (*ColumnSchema, error) {
	f, ok := ts.name2Column[name]
	if ok {
		return f, nil
	}
	return nil, errors.New("not exist")
}

type ColumnSchema struct {
	Name          string // 字段名
	IsPrimaryKey  bool
	AutoIncrement bool // 自增长
	Kind          dbcodec.Kind
	Unique        bool         // 是否唯一键
	Index         *IndexSchema // 索引的名称
	UniqueIndex   *IndexSchema // 唯一索引
	Size          int          // 定义列数据类型的大小或长度
	NotNull       bool
	Codec         string // 字段编解码器

	Default *DefaultValueSchema
}

type IndexSchema struct {
	FieldName  string
	IndexName  string // 索引
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

func Schema(obj any) (*TableSchema, error) {
	return (schemaParser{}).Parser(obj)
}

type hasTable interface {
	TableName() string
}

type schemaParser struct{}

func (sp schemaParser) Parser(obj any) (*TableSchema, error) {
	rt := reflect.TypeOf(obj)
	if rt.Kind() != reflect.Struct {
		return nil, errors.New("obj is not a struct")
	}
	var table string
	if ht, ok := obj.(hasTable); ok {
		table = ht.TableName()
	}
	value := schemaCache.Get2(rt, sp.getSchemaCacheValue)
	if value.Err != nil {
		return nil, value.Err
	}
	sc := value.Schema
	sc.Table = table
	return &sc, nil
}

var schemaCache = zcache.MapCache[reflect.Type, *schemaCacheValue]{}

type schemaCacheValue struct {
	Schema TableSchema
	Err    error
}

func (sp schemaParser) getSchemaCacheValue(rt reflect.Type) *schemaCacheValue {
	sc, err := sp.getSchema(rt)
	return &schemaCacheValue{
		Schema: *sc,
		Err:    err,
	}
}

func (sp schemaParser) getSchema(rt reflect.Type) (*TableSchema, error) {
	sc := &TableSchema{
		name2Column: make(map[string]*ColumnSchema),
	}

	var scan func(reflect.Type) error

	tn := TagName()
	scan = func(rt reflect.Type) error {
		err := zreflect.RangeStructFields(rt, func(field reflect.StructField) error {
			// embed 类型的，详见 testUser3、testUser4
			if field.Anonymous {
				switch field.Type.Kind() {
				case reflect.Struct:
					return scan(field.Type)
				case reflect.Ptr:
					return scan(field.Type.Elem())
				default:
					panic(fmt.Sprintf("what Anonymous %s", field.Type.Kind()))
				}
			}

			if !field.IsExported() {
				return nil
			}

			tag := xstruct.ParserTagCached(field.Tag, tn)
			name := tag.Name()
			if name == "-" || name == "" {
				return nil
			}
			scf, err := sp.parserField(field, tag)
			if err != nil {
				return fmt.Errorf("field=%q: %w", field.Name, err)
			}
			if _, has := sc.name2Column[name]; has {
				return fmt.Errorf("struct Field %q has duplicate column %q", field.Name, name)
			}
			sc.name2Column[name] = scf
			sc.Columns = append(sc.Columns, scf)
			return nil
		})
		return err
	}

	err := scan(rt)
	return sc, err
}

func (sp schemaParser) parserField(f reflect.StructField, tag xstruct.Tag) (*ColumnSchema, error) {
	field := &ColumnSchema{
		Name:          tag.Name(),
		AutoIncrement: TagHasAutoIncr(tag),
		IsPrimaryKey:  TagHasPrimaryKey(tag),
		NotNull:       tag.Has(TagNotNull),
		Unique:        tag.Has(TagUnique),
	}

	var err error
	field.Index, err = sp.parserIndex(field.Name, tag, false)
	if err == nil {
		field.UniqueIndex, err = sp.parserIndex(field.Name, tag, true)
	}
	if err != nil {
		return nil, err
	}

	if size, has := tag.Get(TagSize); has {
		num, err0 := strconv.Atoi(size)
		if err0 != nil || num <= 0 {
			return nil, fmt.Errorf("invalid size: %s", size)
		}
		field.Size = num
	}

	tp := tag.Value(TagType)
	if tp != "" {
		tk := dbcodec.Kind(tp)
		if !tk.IsValid() {
			return nil, fmt.Errorf("invalid type: %q", tp)
		}
		field.Kind = tk
	}

	codec := tag.Value(TagCodec)
	if codec != "" {
		field.Codec = codec
		cc, err0 := dbcodec.Find(codec)
		if err0 != nil {
			return nil, err0
		}
		if !field.Kind.IsValid() {
			field.Kind = cc.Kind()
		}
	}
	if !field.Kind.IsValid() {
		field.Kind, err = dbcodec.ReflectToKind(f.Type)
	}

	if err != nil {
		return nil, err
	}
	if def, ok := tag.Get(TagDefault); ok {
		field.Default, err = sp.parserDefault(def)
	}
	return field, err
}

func (sp schemaParser) parserIndex(fieldName string, tag xstruct.Tag, isUniq bool) (*IndexSchema, error) {
	indexTagName := TagIndex
	indexNamePrefix := "idx_"
	if isUniq {
		indexTagName = TagUniqueIndex
		indexNamePrefix += "uniq_"
	}
	index, has := tag.Get(indexTagName)
	if !has {
		return nil, nil
	}

	if index == "" {
		return &IndexSchema{
			IndexName:  indexNamePrefix + fieldName,
			FieldOrder: -1,
		}, nil
	}

	idxName, order, found := strings.Cut(index, ",")
	if found {
		num, err := strconv.Atoi(order)
		if err != nil || num < 0 {
			return nil, fmt.Errorf("invalid field order in: %q", order)
		}
		return &IndexSchema{
			IndexName:  idxName,
			FieldOrder: num,
		}, nil
	}

	return &IndexSchema{
		IndexName:  index,
		FieldOrder: 0,
	}, nil
}

var regNumber = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

func (sp schemaParser) parserDefault(def string) (*DefaultValueSchema, error) {
	def = strings.TrimSpace(def)
	if def == "" {
		return &DefaultValueSchema{
			Type:  DefaultValueTypeString,
			Value: "",
		}, nil
	}
	tp, val, found := strings.Cut(def, "|")
	tp = strings.TrimSpace(tp)
	val = strings.TrimSpace(val)
	if !found {
		return &DefaultValueSchema{
			Type:  DefaultValueTypeString,
			Value: val,
		}, nil
	}
	switch tp {
	case "number":
		if !regNumber.MatchString(val) {
			return nil, fmt.Errorf("invalid number: %q", val)
		}
		return &DefaultValueSchema{
			Type:  DefaultValueTypeNumber,
			Value: val,
		}, nil
	case "fn":
		return &DefaultValueSchema{
			Type:  DefaultValueTypeFn,
			Value: val,
		}, nil
	case "string":
		return &DefaultValueSchema{
			Type:  DefaultValueTypeString,
			Value: val,
		}, nil
	default:
		return nil, fmt.Errorf("invalid default value %q", def)
	}
}
