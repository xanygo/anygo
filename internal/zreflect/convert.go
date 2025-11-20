//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-04

package zreflect

import (
	"fmt"
	"reflect"
	"strconv"
)

// BaseTypeToString 将基础类型转换为字符串
func BaseTypeToString(va any) (string, bool) {
	rv := reflect.ValueOf(va)
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(rv.Uint(), 10), true

	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, rv.Type().Bits()), true
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool()), true
	case reflect.Pointer:
		return BaseTypeToString(rv.Elem().Interface())
	default:
		return "", false
	}
}

func ToString(va any) string {
	vs, ok := BaseTypeToString(va)
	if ok {
		return vs
	}
	return fmt.Sprint(va)
}

// BaseTypeToInt64 将基础类型转换为 int64
func BaseTypeToInt64(va any) (int64, bool) {
	rv := reflect.ValueOf(va)
	switch rv.Kind() {
	case reflect.String:
		num, err := strconv.ParseInt(rv.String(), 10, 64)
		return num, err == nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(rv.Uint()), true

	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	case reflect.Pointer:
		return BaseTypeToInt64(rv.Elem().Interface())
	case reflect.Bool:
		if rv.Bool() {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func IsBasicKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

func BytesValue(obj any) ([]byte, bool) {
	if obj == nil {
		return nil, false
	}
	v := reflect.ValueOf(obj)

	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
		return v.Bytes(), true
	}

	return nil, false
}
