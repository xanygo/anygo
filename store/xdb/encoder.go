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
	"github.com/xanygo/anygo/store/xdb/dbcodec"
)

// Encode 将结构体或 map 转成 map[string]any，用于 SQL insert
func Encode(data any, cols ...string) (map[string]any, error) {
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
		return encodeStruct(v, cols)
	case reflect.Map:
		return encodeMap(v, cols)
	default:
		return nil, fmt.Errorf("unsupported type %T", data)
	}
}

func EncodeBatch[T any](items []T, fields ...string) ([]map[string]any, error) {
	if len(items) == 0 {
		return nil, errors.New("no value to encode")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data, err := Encode(item, fields...)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}
	return result, nil
}

// encodeStruct 处理 struct
func encodeStruct(v reflect.Value, fields []string) (map[string]any, error) {
	t := v.Type()
	result := make(map[string]any)
	fieldsLimit := xslice.ToMap(fields, true)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !v.Field(i).CanInterface() {
			continue
		}
		val := v.Field(i).Interface()

		tag := xstruct.ParserTag(field.Tag.Get(TagName()))
		name := tag.Name()
		if name == "-" || name == "" {
			continue
		}
		if _, has := result[name]; has {
			return nil, fmt.Errorf("duplicate field %q", name)
		}

		if len(fieldsLimit) > 0 && !fieldsLimit[name] {
			continue
		}

		encodedVal, err := encodeStructFieldValue(val, tag)
		if err != nil {
			return nil, fmt.Errorf("encode field %q: %w", field.Name, err)
		}
		result[name] = encodedVal
	}

	return result, nil
}

// encodeStructFieldValue 对单个字段根据类型和 serializer 转换
func encodeStructFieldValue(val any, tag xstruct.Tag) (any, error) {
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

	if zreflect.IsBasicKind(rv.Kind()) {
		return val, nil
	}
	// 对 slice / map / struct / pointer 类型用 serializer
	codec := getCodecName(tag)
	encoder, err := dbcodec.Find(codec)
	if err != nil {
		return nil, err
	}
	return encoder.Encode(val)
}

// encodeMap 处理 map[string]any
func encodeMap(v reflect.Value, fields []string) (map[string]any, error) {
	fieldsLimit := xslice.ToMap(fields, true)
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
		result[key] = val
	}
	return result, nil
}

type Expr string
