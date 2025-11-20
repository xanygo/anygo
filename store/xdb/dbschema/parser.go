//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbschema

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zcache"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// Schema 传入一个 struct，获取其 db schema 定义
func Schema(fy dbtype.Dialect, obj any) (*dbtype.TableSchema, error) {
	return (schemaParser{fy: fy}).Parser(obj)
}

type hasTable interface {
	TableName() string
}

type schemaParser struct {
	fy dbtype.Dialect
}

func (sp schemaParser) Parser(obj any) (*dbtype.TableSchema, error) {
	rt := reflect.TypeOf(obj)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
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
	Schema dbtype.TableSchema
	Err    error
}

func (sp schemaParser) getSchemaCacheValue(rt reflect.Type) *schemaCacheValue {
	sc, err := sp.getSchema(rt)
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
			if _, has := sc.Name2Column[name]; has {
				return fmt.Errorf("struct Field %q has duplicate column %q", field.Name, name)
			}
			sc.Name2Column[name] = scf
			sc.Columns = append(sc.Columns, scf)
			return nil
		})
		return err
	}

	err := scan(rt)
	sc.ColumnsNames = xmap.Keys(sc.Name2Column)
	slices.Sort(sc.ColumnsNames)
	return sc, err
}

func (sp schemaParser) parserField(f reflect.StructField, tag xstruct.Tag) (dbtype.ColumnSchema, error) {
	field := dbtype.ColumnSchema{
		ReflectType: f.Type,

		Name:          tag.Name(),
		AutoIncrement: TagHasAutoInc(tag),
		IsPrimaryKey:  TagHasPrimaryKey(tag),
		NotNull:       tag.Has(TagNotNull),
		Unique:        TagHasUnique(tag),
		Native:        tag.Value(TagNative),
	}

	var err error
	field.Index, err = sp.parserIndex(field.Name, tag, false)
	if err == nil {
		field.UniqueIndex, err = sp.parserIndex(field.Name, tag, true)
	}
	if err != nil {
		return field, err
	}

	if size, has := tag.Get(TagSize); has {
		num, err0 := strconv.Atoi(size)
		if err0 != nil || num <= 0 {
			return field, fmt.Errorf("invalid size: %s", size)
		}
		field.Size = num
	}

	tp := tag.Value(TagType)
	if tp != "" {
		tk := dbtype.Kind(tp)
		if !tk.IsValid() {
			return field, fmt.Errorf("invalid type: %q", tp)
		}
		field.Kind = tk
	}

	codecName := tag.Value(TagCodec)
	if codecName != "" && !dbtype.KindAutoJSON.Is(codecName) {
		field.Codec, err = findCodec(sp.fy, codecName)
		if err != nil {
			return field, err
		}
	}

	if field.Codec == nil {
		if dz, ok := sp.fy.(dbtype.CoderDialect); ok {
			field.Codec, err = dz.ColumnCodec(f.Type)
			if err != nil {
				return field, err
			}
		}
	}

	if field.Codec != nil && !field.Kind.IsValid() {
		// 当有明确的 Codec 的时候，使用 Codec 的 Kind
		field.Kind = field.Codec.Kind()
	}

	if field.Codec == nil {
		if codecName != "" && dbtype.KindAutoJSON.Is(codecName) {
			field.Codec = dbcodec.JSON{}
		} else {
			field.Codec, err = findCodec(sp.fy, dbcodec.TextName)
			if err != nil {
				return field, err
			}
		}
	}

	if !field.Kind.IsValid() {
		field.Kind, err = dbtype.ReflectToKind(f.Type)
	}

	if err != nil {
		return field, err
	}
	if def, ok := tag.Get(TagDefault); ok {
		field.Default, err = sp.parserDefault(def)
	}
	return field, err
}

func findCodec(d dbtype.Dialect, name string) (dbtype.Codec, error) {
	return dbcodec.Find(name+"@"+d.Name(), name)
}

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
			IndexName:  idxName,
			FieldOrder: num,
		}, nil
	}

	return &dbtype.IndexSchema{
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
