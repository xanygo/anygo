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

	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zcache"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// Schema 传入一个 struct，获取其 db schema 定义.
//
// schema 的 tag 字段来自于：
//  1. 若 obj 有实现 type xxx interface{  XDBTag()string  }
//  2. 使用  TagName()，默认值为 "db"
func Schema(fy dbtype.Dialect, obj any) (*dbtype.TableSchema, error) {
	tn := zreflect.CallStringMethod(obj, "XDBTag")
	if tn == "" {
		tn = TagName()
	}
	return (schemaParser{dialect: fy, tagName: tn}).Parser(obj)
}

type schemaParser struct {
	dialect dbtype.Dialect
	tagName string // 在 struct 里定义的 db 的 tag 名称
}

func (sp schemaParser) Parser(obj any) (*dbtype.TableSchema, error) {
	rt := reflect.TypeOf(obj)

	value := schemaCache.Get2(schemaCacheKey{Type: rt, Dialect: sp.dialect.Name()}, func(key schemaCacheKey) *schemaCacheValue {
		return sp.parserFromCache(key, obj)
	})
	if value.Err != nil {
		return nil, value.Err
	}
	sc := value.Schema
	return &sc, nil
}

func (sp schemaParser) parserFromCache(key schemaCacheKey, obj any) *schemaCacheValue {
	cv := sp.getSchemaCacheValue(key)
	if cv.Err != nil {
		return cv
	}
	cv.Schema.Table = zreflect.CallStringMethod(obj, "TableName")
	return cv
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
		TagName:     sp.tagName,
		Name2Column: make(map[string]dbtype.ColumnSchema),
	}

	raw := rt
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return sc, fmt.Errorf("dbschema: invalid type %T, should struct or *struct", raw.String())
	}

	var scan func(reflect.Type) error

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

			tag := xstruct.ParserTagCached(field.Tag, sp.tagName)
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
		ReflectType:   ft,
		Name:          tag.Name(),
		AutoIncrement: TagHasAutoInc(tag),
		IsPrimaryKey:  TagHasPrimaryKey(tag),
		Unique:        TagHasUnique(tag),
		NotNull:       !tag.Has(TagNull),
		Native:        tag.Value(TagNative),
		Kind:          dbtype.Kind(tag.Value(TagType)),
		Auto:          tag.Value(TagAuto),
		Group:         xstr.ToStrings(tag.Value(TagGroup), ","),
	}

	if field.Kind != "" && !field.Kind.IsOK() {
		return field, fmt.Errorf("invalid type: %q", field.Kind)
	}

	var err error
	field.Indexes, err = sp.parserIndexes(field.Name, tag)
	if err != nil {
		return field, err
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

var indexReg = regexp.MustCompile(`^(\w+)(?:\[(\d+)\])?$`)

// parserIndex 解析定义的索引字段
func (sp schemaParser) parserIndexes(fieldName string, tag xstruct.Tag) ([]*dbtype.IndexSchema, error) {
	var result []*dbtype.IndexSchema
	parserOne := func(tagName string, isUniq bool, prefix string) error {
		str, has := tag.Get(tagName)
		if !has {
			return nil
		}
		// `db:"name,index"` 或者  `db:"name,uniq_index"`
		if str == "" {
			index := &dbtype.IndexSchema{
				Unique:     isUniq,
				FieldName:  fieldName,
				IndexName:  prefix + fieldName,
				FieldOrder: -1,
			}
			result = append(result, index)
			return nil
		}
		// `db:"name,index=idx_name"` 或者  `db:"name,uniq_index=name"`
		// `db:"name,index=idx_name[2]"` 或者  `db:"name,uniq_index=name[2]"`
		// `db:"name,index=idx_name[2];index_name_class"` 或者  `db:"name,uniq_index=name;name_class"`
		arr := strings.Split(str, ";")
		for _, subStr := range arr {
			subStr = strings.TrimSpace(subStr)
			if subStr == "" {
				continue
			}
			matches := indexReg.FindStringSubmatch(subStr)
			if len(matches) == 0 {
				return fmt.Errorf("invalid index: %q", subStr)
			}
			var order int
			if matches[2] != "" {
				num, err := strconv.Atoi(matches[2])
				if err != nil || num < 0 {
					return fmt.Errorf("invalid field order in: %q", subStr)
				}
				order = num
			}
			index := &dbtype.IndexSchema{
				FieldName:  fieldName,
				Unique:     isUniq,
				IndexName:  prefix + matches[1],
				FieldOrder: order,
			}
			result = append(result, index)
		}

		return nil
	}

	if err := parserOne(TagIndex, false, "idx_"); err != nil {
		return nil, err
	}
	if err := parserOne(TagUniqueIndex, true, "uniq_"); err != nil {
		return nil, err
	}
	// 检查是否有重名的
	names := make(map[string]struct{}, len(result))
	for _, index := range result {
		if _, ok := names[index.IndexName]; ok {
			return nil, fmt.Errorf("duplicate index: %q", index.IndexName)
		}
		names[index.IndexName] = struct{}{}
	}
	return result, nil
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
