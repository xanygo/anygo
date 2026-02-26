//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xdb

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
)

// Encode 将结构体或 map 转成 map[string]any，用于 SQL insert
func Encode[T any](fy dbtype.Dialect, data T, cols ...string) (map[string]any, error) {
	return Encoder[T]{Dialect: fy, OnlyFields: cols}.Encode(data)
}

func EncodeBatch[T any](fy dbtype.Dialect, items []T, fields ...string) ([]map[string]any, error) {
	if len(items) == 0 {
		return nil, errors.New("no value to encode")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data, err := Encode(fy, item, fields...)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}
	return result, nil
}

type Encoder[T any] struct {
	Dialect      dbtype.Dialect
	OnlyFields   []string // 当不为空时，输出结果的 key 只限定此列表中的
	IgnoreFields []string // 当不为空时，输出结果的 key 限定不能是此列表中的
}

func (e Encoder[T]) Encode(data T) (map[string]any, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid value: %v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer: %#v", data)
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		schema, err := dbschema.Schema(e.Dialect, data)
		if err != nil {
			return nil, err
		}
		return e.encodeStruct(v, schema)
	case reflect.Map:
		return e.encodeMap(v)
	default:
		return nil, fmt.Errorf("unsupported type %T", data)
	}
}

func (e Encoder[T]) EncodeBatch(items ...T) ([]map[string]any, error) {
	if len(items) == 0 {
		return nil, errors.New("no value to encode")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data, err := e.Encode(item)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}
	return result, nil
}

// encodeStruct 处理 struct
func (e Encoder[T]) encodeStruct(v reflect.Value, schema *dbtype.TableSchema) (map[string]any, error) {
	result := make(map[string]any, len(e.OnlyFields))
	err := e.withStruct(v, func(name string, tag xstruct.Tag, field reflect.StructField, value reflect.Value) error {
		fs, err := schema.ColumnByName(name)
		if err != nil {
			return err
		}
		encodedVal, err := e.encodeStructFieldValue(fs, value.Interface())
		if err != nil {
			return fmt.Errorf("encode field %q: %w", field.Name, err)
		}
		result[name] = encodedVal
		return nil
	})
	return result, err
}

var sliceEmpty = map[string]bool{}

func (e Encoder[T]) sliceToMapTrue(s []string) map[string]bool {
	if len(s) == 0 {
		return sliceEmpty
	}
	return xslice.ToMap(e.OnlyFields, true)
}

func (e Encoder[T]) withStruct(v reflect.Value, fn func(name string, tag xstruct.Tag, field reflect.StructField, value reflect.Value) error) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type %s, expect struct", v.Kind().String())
	}
	keys := make(map[string]struct{}, len(e.OnlyFields))
	fieldsLimit := e.sliceToMapTrue(e.OnlyFields)
	fieldsIgnore := e.sliceToMapTrue(e.IgnoreFields)
	tn := dbschema.TagName()
	err := zreflect.RangeStructFields(v.Type(), func(field reflect.StructField) error {
		// embed 类型的，详见 testUser3、testUser4
		if field.Anonymous {
			fv := v.FieldByName(field.Name)
			switch fv.Kind() {
			case reflect.Struct:
				return e.withStruct(fv, fn)
			case reflect.Pointer:
				return e.withStruct(fv.Elem(), fn)
			default:
				panic(fmt.Sprintf("what Anonymous %s: %v", fv.Kind(), fv))
			}
		}
		if !field.IsExported() {
			return nil
		}
		fv := v.FieldByName(field.Name)
		if !fv.CanInterface() {
			return nil
		}

		tag := xstruct.ParserTagCached(field.Tag, tn)
		name := tag.Name()
		if name == "-" || name == "" {
			return nil
		}
		if _, has := keys[name]; has {
			return fmt.Errorf("duplicate field %q", name)
		}

		if len(fieldsLimit) > 0 && !fieldsLimit[name] {
			return nil
		}
		if len(fieldsIgnore) > 0 && fieldsIgnore[name] {
			return nil
		}
		if dbschema.TagHasAutoInc(tag) && fv.IsZero() {
			// 当时自增长类型、并且是零值，则跳过该字段
			return nil
		}
		if err := fn(name, tag, field, fv); err != nil {
			return err
		}
		keys[name] = struct{}{}
		return nil
	})
	return err
}

// encodeMap 处理 map[string]any
func (e Encoder[T]) encodeMap(v reflect.Value) (map[string]any, error) {
	fieldsLimit := e.sliceToMapTrue(e.OnlyFields)
	fieldsIgnore := e.sliceToMapTrue(e.IgnoreFields)
	result := make(map[string]any)
	for _, k := range v.MapKeys() {
		val := v.MapIndex(k).Interface()
		if k.Kind() != reflect.String {
			return nil, fmt.Errorf("key %#v is not a string", val)
		}
		key := k.String()
		if len(fieldsLimit) > 0 && !fieldsLimit[key] {
			continue
		}
		if len(fieldsIgnore) > 0 && fieldsIgnore[key] {
			continue
		}
		result[key] = val
	}
	return result, nil
}

// Fields 获取 data 的字段列表。
//
//  1. 当类型是 struct 的时候，返回有有效的 db tag的字段（以使用 OnlyFiles 和 IgnoreFields 过滤）。
//     若返回的字段列表为空，会报错。
//  2. 当类型是 map 时，返回 nil,nil
//  3. 其他类型，返回 error
func (e Encoder[T]) Fields(data T) ([]string, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid value: %v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer: %#v", data)
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return e.structFields(data)
	case reflect.Map:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", data)
	}
}

func (e Encoder[T]) structFields(v T) ([]string, error) {
	sc, err := dbschema.Schema(e.Dialect, v)
	if err != nil {
		return nil, err
	}
	return sc.ColumnsNames, nil
}

// encodeStructFieldValue 对单个字段根据类型和 serializer 转换
func (e Encoder[T]) encodeStructFieldValue(schema dbtype.ColumnSchema, val any) (any, error) {
	rv := reflect.ValueOf(val)
	if !rv.IsValid() {
		return nil, fmt.Errorf("invalid value: %v", val)
	}
	// 处理指针
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv = reflect.New(rv.Type().Elem()).Elem()
			val = rv.Interface()
		} else {
			rv = rv.Elem()
			val = rv.Interface()
		}
	}

	// 优先查找方言里有没有针对类型的 codec
	if dc, ok := e.Dialect.(dbtype.CoderDialect); ok {
		if encoder, err := dc.ColumnCodec(rv.Type()); err != nil {
			return nil, err
		} else if encoder != nil {
			return encoder.Encode(val)
		}
	}

	if schema.Kind == dbtype.KindNative || zreflect.IsBasicKind(rv.Kind()) {
		return val, nil
	}

	// 对 slice / map / struct / pointer 类型用 serializer
	if schema.Codec == nil {
		return val, nil
	}
	return schema.Codec.Encode(val)
}

// PKNameAndValue 查找主键 key，并返回其值，并且要求值为非零值。
func (e Encoder[T]) PKNameAndValue(obj T) (name string, value any, err error) {
	cs, fv, err := e.PrimaryKey(obj)
	if err != nil {
		return "", nil, err
	}
	if fv.IsZero() {
		return "", nil, fmt.Errorf("pk(%q) is zero value", cs.Name)
	}
	value, err = e.encodeStructFieldValue(cs, fv.Interface())
	if err != nil {
		return "", nil, err
	}
	return cs.Name, value, err
}

func (e Encoder[T]) PrimaryKey(obj T) (cs dbtype.ColumnSchema, fv reflect.Value, err error) {
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		return cs, fv, fmt.Errorf("invalid value: %v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return cs, fv, fmt.Errorf("nil pointer: %#v", obj)
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return cs, fv, fmt.Errorf("invalid value: %#v", obj)
	}

	schema, err := dbschema.Schema(e.Dialect, obj)
	if err != nil {
		return cs, fv, err
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !v.Field(i).CanInterface() {
			continue
		}
		fv = v.Field(i)

		tag := xstruct.ParserTag(field.Tag.Get(dbschema.TagName()))
		name := tag.Name()
		if name == "-" || name == "" {
			continue
		}
		cs, err = schema.ColumnByName(name)
		if err != nil {
			return cs, fv, err
		}
		if !cs.IsPrimaryKey {
			continue
		}
		return cs, fv, nil
	}

	return cs, fv, fmt.Errorf("no primary key column found for %T", obj)
}
