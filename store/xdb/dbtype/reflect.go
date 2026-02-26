//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package dbtype

import (
	"fmt"
	"reflect"
	"time"
)

var typeToKindMap = map[reflect.Kind]Kind{
	reflect.Int:   KindInt,
	reflect.Int8:  KindInt8,
	reflect.Int16: KindInt16,
	reflect.Int32: KindInt32,
	reflect.Int64: KindInt64,

	reflect.Uint:   KindUint,
	reflect.Uint8:  KindUint8,
	reflect.Uint16: KindUint16,
	reflect.Uint32: KindUint32,
	reflect.Uint64: KindUint64,

	reflect.Float32: KindFloat32,
	reflect.Float64: KindFloat64,

	reflect.String: KindString,

	reflect.Bool: KindBoolean,

	reflect.Struct:    KindJSON,
	reflect.Map:       KindJSON,
	reflect.Slice:     KindJSON,
	reflect.Array:     KindJSON,
	reflect.Interface: KindJSON,
}

var specTypeToKindMap = map[reflect.Type]Kind{
	reflect.TypeFor[time.Time](): KindDateTime,
	reflect.TypeFor[[]byte]():    KindBinary,
}

func ReflectToKind(rt reflect.Type) (Kind, error) {
	if k, ok := specTypeToKindMap[rt]; ok {
		return k, nil
	}
	kind := rt.Kind()
	if k, ok := typeToKindMap[kind]; ok {
		return k, nil
	}
	if kind == reflect.Pointer {
		return ReflectToKind(rt.Elem())
	}

	return KindInvalid, fmt.Errorf("invalid data type: %s", rt.String())
}
