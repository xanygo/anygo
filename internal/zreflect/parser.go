//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package zreflect

import (
	"fmt"
	"reflect"
	"strconv"
)

// ParseBasicValue 字符串解析为指定的基础类型
func ParseBasicValue(s string, typ reflect.Type) (reflect.Value, error) {
	switch typ.Kind() {
	case reflect.String:
		return reflect.ValueOf(s).Convert(typ), nil
	case reflect.Int:
		i, err := strconv.Atoi(s)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil
	case reflect.Int8:
		i, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil
	case reflect.Int16:
		i, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil

	case reflect.Uint:
		u, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(u).Convert(typ), nil
	case reflect.Uint8:
		u, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(u).Convert(typ), nil
	case reflect.Uint16:
		u, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(u).Convert(typ), nil
	case reflect.Uint32:
		u, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(u).Convert(typ), nil
	case reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(u).Convert(typ), nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(f).Convert(typ), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(f).Convert(typ), nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(b).Convert(typ), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported element type: %v", typ)
	}
}
