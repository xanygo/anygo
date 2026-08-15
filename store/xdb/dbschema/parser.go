//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbschema

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zcache"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// Schema 传入一个 struct，获取其 db schema 定义
func Schema(fy dbtype.Dialect, obj any) (*dbtype.TableSchema, error) {
	return (schemaParser{dialect: fy}).Parser(obj)
}

type hasTable interface {
	TableName() string
}

type schemaParser struct {
	dialect dbtype.Dialect
}

func (sp schemaParser) Parser(obj any) (*dbtype.TableSchema, error) {
	rt := reflect.TypeOf(obj)
	isPtr := rt.Kind() == reflect.Pointer
	if isPtr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dbschema: invalid type %T, should struct or *struct", obj)
	}

	value := schemaCache.Get2(schemaCacheKey{Type: rt, Dialect: sp.dialect.Name()}, func(key schemaCacheKey) *schemaCacheValue {
		cv := sp.getSchemaCacheValue(key)
		if cv.Err != nil {
			return cv
		}
		var table string
		if ht, ok := obj.(hasTable); ok {
			table = ht.TableName()
		}
		if table == "" && !isPtr {
			rv := reflect.ValueOf(obj)
			ptr := reflect.New(rv.Type())
			ptr.Elem().Set(rv)
			if ht, ok := ptr.Interface().(hasTable); ok {
				table = ht.TableName()
			}
		}
		cv.Schema.Table = table
		return cv
	})
	if value.Err != nil {
		return nil, value.Err
	}
	sc := value.Schema
	return &sc, nil
}

type schemaCacheKey struct {
	Type    reflect.Type
	Dialect string
}

var schemaCache = zcache.MapCache[schemaCacheKey, *schemaCacheValue]{}

type schemaCacheValue struct {
	Schema dbtype.TableSchema
	Err    error
}

func (sp schemaParser) getSchemaCacheValue(key schemaCacheKey) *schemaCacheValue {
	sc, err := sp.getSchema(key.Type)
	return &schemaCacheValue{
		Schema: *sc,
		Err:    err,
	}
}

func (sp schemaParser) getSchema(rt reflect.Type) (*dbtype.TableSchema, error) {
	sc := &dbtype.TableSchema{
		Name2Column: make(map[string]dbtype.ColumnSchema),
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
				case reflect.Pointer:
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
			if _, has := sc.Name2Column[name]; has {
				return fmt.Errorf("struct Field %q has duplicate column %q", field.Name, name)
			}
			sc.Name2Column[name] = scf
			sc.Columns = append(sc.Columns, scf)
			sc.ColumnNames = append(sc.ColumnNames, name)
			return nil
		})
		return err
	}

	err := scan(rt)
	return sc, err
}

func (sp schemaParser) parserField(f reflect.StructField, tag xstruct.Tag) (dbtype.ColumnSchema, error) {
	ft := f.Type
	field := dbtype.ColumnSchema{
		ReflectType: ft,

		Name:          tag.Name(),
		AutoIncrement: TagHasAutoInc(tag),
		IsPrimaryKey:  TagHasPrimaryKey(tag),
		NotNull:       tag.Has(TagNotNull),
		Unique:        TagHasUnique(tag),
		Native:        tag.Value(TagNative),
		Kind:          dbtype.Kind(tag.Value(TagType)),
	}

	if field.Kind != "" && !field.Kind.IsOK() {
		return field, fmt.Errorf("invalid type: %q", field.Kind)
	}

	var err error
	field.Index, err = sp.parserIndex(field.Name, tag, false)
	if err == nil {
		field.UniqueIndex, err = sp.parserIndex(field.Name, tag, true)
	}

	if def, ok := tag.Get(TagDefault); ok {
		field.Default, err = sp.parserDefault(def)
	}

	if err != nil {
		return field, err
	}

	// 解析 size 字段属性
	{
		if size, has := tag.Get(TagSize); has {
			num, err0 := strconv.Atoi(size)
			if err0 != nil || num <= 0 {
				return field, fmt.Errorf("invalid size: %s", size)
			}
			field.Size = num
		}

		// 获取数组 [N]byte 的长度
		if field.Size == 0 && ft.Kind() == reflect.Array && ft.Elem().Kind() == reflect.Uint8 {
			field.Size = ft.Len()
		}
	}

	if err = sp.parserCodec(&field, f, tag); err != nil {
		return field, err
	}

	return field, err
}

func (sp schemaParser) trySetKindByCodec(field *dbtype.ColumnSchema) {
	if field.Kind.IsOK() {
		return
	}
	if hc, ok := field.Codec.(dbtype.HasKind); ok {
		field.Kind = hc.Kind()
	}
}

func (sp schemaParser) parserCodec(field *dbtype.ColumnSchema, f reflect.StructField, tag xstruct.Tag) (err error) {
	codecName := tag.Value(TagCodec)
	if codecName != "" && codecName != codecAutoJSON {
		codec := findCodec(sp.dialect, codecName)
		if codec == nil {
			return fmt.Errorf("invalid codec %q", codecName)
		}
		field.Codec = codec

		sp.trySetKindByCodec(field)
	}

	if dz, ok := sp.dialect.(dbtype.CoderDialect); ok {
		kind, codec, native := dz.ColumnCodec(f.Type)
		if !field.Kind.IsOK() {
			field.Kind = kind
		}
		if field.Codec == nil {
			field.Codec = codec
		}
		if field.Native == "" {
			field.Native = native
		}
	}

	if !field.Kind.IsOK() {
		field.Kind, _ = dbtype.ReflectToKind(f.Type)
	}

	if field.Codec == nil && field.Kind == dbtype.KindBinary {
		field.Codec = dbcodec.Binary{}
	}

	if field.Codec == nil && codecName == codecAutoJSON {
		field.Codec = dbcodec.JSON{}
	}

	sp.trySetKindByCodec(field)

	if field.Codec == nil && field.Kind.IsOK() {
		field.Codec = dbcodec.FindByKind(field.Kind)
	}

	if field.Codec == nil {
		if zreflect.IsBasicKind(f.Type.Kind()) {
			field.Codec = dbcodec.Native{}
		} else {
			field.Codec = findCodec(sp.dialect, dbcodec.TextName)
		}
	}
	return nil
}

func findCodec(d dbtype.Dialect, name string) dbtype.Codec {
	return dbcodec.FindByName(name+"@"+d.Name(), name)
}

// parserIndex 解析定义的索引字段
func (sp schemaParser) parserIndex(fieldName string, tag xstruct.Tag, isUniq bool) (*dbtype.IndexSchema, error) {
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
		return &dbtype.IndexSchema{
			FieldName:  fieldName,
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
		return &dbtype.IndexSchema{
			FieldName:  fieldName,
			IndexName:  idxName,
			FieldOrder: num,
		}, nil
	}

	return &dbtype.IndexSchema{
		FieldName:  fieldName,
		IndexName:  index,
		FieldOrder: 0,
	}, nil
}

var regNumber = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

func (sp schemaParser) parserDefault(def string) (*dbtype.DefaultValueSchema, error) {
	def = strings.TrimSpace(def)
	if def == "" {
		return &dbtype.DefaultValueSchema{
			Type:  dbtype.DefaultValueTypeString,
			Value: "",
		}, nil
	}
	tp, val, found := strings.Cut(def, "|")
	tp = strings.TrimSpace(tp)
	val = strings.TrimSpace(val)
	if !found {
		return &dbtype.DefaultValueSchema{
			Type:  dbtype.DefaultValueTypeString,
			Value: val,
		}, nil
	}
	switch tp {
	case "number":
		if !regNumber.MatchString(val) {
			return nil, fmt.Errorf("invalid number: %q", val)
		}
		return &dbtype.DefaultValueSchema{
			Type:  dbtype.DefaultValueTypeNumber,
			Value: val,
		}, nil
	case "fn":
		return &dbtype.DefaultValueSchema{
			Type:  dbtype.DefaultValueTypeFn,
			Value: val,
		}, nil
	case "string":
		return &dbtype.DefaultValueSchema{
			Type:  dbtype.DefaultValueTypeString,
			Value: val,
		}, nil
	default:
		return nil, fmt.Errorf("invalid default value %q", def)
	}
}
