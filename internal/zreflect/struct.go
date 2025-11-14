//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package zreflect

import (
	"reflect"

	"github.com/xanygo/anygo/internal/zcache"
)

type structMeta struct {
	Fields []reflect.StructField
}

var structMetaCache = &zcache.MapCache[reflect.Type, *structMeta]{}

func loadStructMeta(t reflect.Type) *structMeta {
	v, ok := structMetaCache.Load(t)
	if ok {
		return v
	}
	meta := &structMeta{
		Fields: collectFields(t),
	}
	// 匿名 struct
	if t.Name() != "" && t.PkgPath() != "" {
		structMetaCache.Set(t, meta)
	}
	return meta
}

func collectFields(t reflect.Type) []reflect.StructField {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	var fields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fields = append(fields, f)
	}
	return fields
}

// RangeStructFields 遍历 StructField 或者 StructField 的 Ptr 的 StructField，带有 cache
//
// 相比直接读取，速度快一倍
//
// withCache-4          55358085                20.74 ns/op
// noCache-4            27793435                43.12 ns/op
func RangeStructFields(t reflect.Type, fn func(field reflect.StructField) error) error {
	meta := loadStructMeta(t)
	for _, field := range meta.Fields {
		if err := fn(field); err != nil {
			return err
		}
	}
	return nil
}
