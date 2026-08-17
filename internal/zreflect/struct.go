//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package zreflect

import (
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/internal/zcache"
)

type structMeta struct {
	Fields []reflect.StructField
	Err    error
}

var StructMetaCache = &zcache.MapCache[reflect.Type, *structMeta]{}

func loadStructMeta(t reflect.Type) *structMeta {
	v, ok := StructMetaCache.Load(t)
	if ok {
		return v
	}
	fs, err := collectFields(t)
	meta := &structMeta{
		Fields: fs,
		Err:    err,
	}
	StructMetaCache.Set(t, meta)
	return meta
}

func collectFields(t reflect.Type) ([]reflect.StructField, error) {
	raw := t
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid type %s, not struct or *struct", raw.String())
	}
	var fields []reflect.StructField
	for f := range t.Fields() {
		fields = append(fields, f)
	}
	return fields, nil
}

// RangeStructFields 遍历 struct 或者 *struct 的 Ptr 的 StructField，带有 cache.
// 若是传入的类型错误，会返回 error。
// 只遍历当前 struct 的，不管 StructField 内部的
//
// 相比直接读取，速度快5倍
//
// withCache-4          68107131                17.60 ns/op
// noCache-4            11106657               107.6 ns/op
func RangeStructFields(t reflect.Type, fn func(field reflect.StructField) error) error {
	meta := loadStructMeta(t)
	if meta.Err != nil {
		return meta.Err
	}
	for _, field := range meta.Fields {
		if err := fn(field); err != nil {
			return err
		}
	}
	return nil
}
