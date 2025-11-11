//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xdb

import (
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
)

var tagName = xsync.OnceInit[string]{
	New: func() string {
		return "db"
	},
}

func TagName() string {
	return tagName.Load()
}

func SetTagName(name string) {
	if name == "" {
		panic("empty tag name")
	}
	tagName.Store(name)
}

const (
	tagPrimaryKey = "primaryKey"
	tagCodec      = "codec"
)

func getCodecName(tag xstruct.Tag) string {
	name := tag.Value(tagCodec)
	if name != "" {
		return name
	}
	return dbcodec.TextName
}

func findStructPrimaryKV(obj any) (key string, value any, err error) {
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		return "", nil, fmt.Errorf("invalid value: %v", v)
	}

	// 支持指针类型
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil, fmt.Errorf("nil pointer: %#v", obj)
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("invalid value: %#v", obj)
	}

	t := v.Type()
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
		if !tag.Has(tagPrimaryKey) {
			continue
		}
		if key != "" {
			return "", nil, fmt.Errorf("multiple primary key fields: %s,%s", key, name)
		}
		key = name
		value, err = encodeStructFieldValue(val, tag)
		if err != nil {
			return "", nil, fmt.Errorf("encode field %q: %w", field.Name, err)
		}
	}
	if key == "" {
		return "", nil, fmt.Errorf("no primary key field: %s", t.Name())
	}
	return key, value, nil
}
