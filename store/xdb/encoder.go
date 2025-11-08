//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xdb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/xcodec"
)

// Encode 将结构体或 map 转成 map[string]any，用于 SQL insert
func Encode(data any) (map[string]any, error) {
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
		return encodeStruct(v)
	case reflect.Map:
		return encodeMap(v)
	default:
		return nil, fmt.Errorf("unsupported type %T", data)
	}
}

// encodeStruct 处理 struct
func encodeStruct(v reflect.Value) (map[string]any, error) {
	t := v.Type()
	result := make(map[string]any)

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
		encodedVal, err := encodeField(val, tag)
		if err != nil {
			return nil, fmt.Errorf("encode field %q: %w", field.Name, err)
		}
		result[name] = encodedVal
	}

	return result, nil
}

// encodeMap 处理 map[string]any
func encodeMap(v reflect.Value) (map[string]any, error) {
	result := make(map[string]any)
	for _, k := range v.MapKeys() {
		if k.Kind() != reflect.String {
			continue
		}
		val := v.MapIndex(k).Interface()
		result[k.String()] = val
	}
	return result, nil
}

// encodeField 对单个字段根据类型和 serializer 转换
func encodeField(val any, tag xstruct.Tag) (any, error) {
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

	if isBasicKind(rv.Kind()) {
		return val, nil
	}
	// 对 slice / map / struct / pointer 类型用 serializer
	codec := tag.Value("codec")
	switch codec {
	case "json", "":
		str, err := xcodec.EncodeToString(xcodec.JSON, val)
		if err != nil {
			return str, err
		}
		if str == "null" {
			return "", nil
		}
		return str, nil
	case "csv":
		csvType := rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
		if !csvType || isComplexKind(rv.Type().Elem().Kind()) {
			return nil, fmt.Errorf("invalid codec:csv for type %T", val)
		}
		var parts []string
		for i := 0; i < rv.Len(); i++ {
			parts = append(parts, fmt.Sprint(rv.Index(i).Interface()))
		}
		return strings.Join(parts, ","), nil
	default:
		if cc, ok := codecs[codec]; ok {
			return xcodec.EncodeToString(cc, val)
		}
		return nil, fmt.Errorf("unsupported codec: %s", codec)
	}
}
