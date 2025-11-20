//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xcodec

import (
	"encoding"
	"fmt"
	"reflect"
	"slices"

	"github.com/xanygo/anygo/internal/zreflect"
)

var _ Codec = (*TextCodec)(nil)

type TextCodec struct{}

func (t TextCodec) Name() string {
	return "text"
}

func (t TextCodec) Encode(obj any) ([]byte, error) {
	if mt, ok := obj.(encoding.TextMarshaler); ok {
		return mt.MarshalText()
	}
	if str, ok := zreflect.BaseTypeToString(obj); ok {
		return []byte(str), nil
	}
	if b, ok := zreflect.BytesValue(obj); ok {
		return b, nil
	}
	return nil, fmt.Errorf("type %T not implement TextMarshaler", obj)
}

// Decode 将bytes 解析到 obj，具体规则如下：
//
//  1. obj 实现了 TextUnmarshaler ，则优先使用
//  2. 若 obj 是  string 或者 []byte 类型，则直接赋值
//  3. 若 obj 是基础类型，如 number、bool 类型，则尝试解析赋值
//  4. 返回错误
func (t TextCodec) Decode(bytes []byte, obj any) error {
	mt, ok := obj.(encoding.TextUnmarshaler)
	if ok {
		return mt.UnmarshalText(bytes)
	}

	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("obj must be a non-nil pointer, got %T", obj)
	}
	elem := rv.Elem()
	kind := elem.Kind()
	switch kind {
	case reflect.String:
		elem.SetString(string(bytes))
		return nil
	case reflect.Slice, reflect.Array:
		if elem.Type().Elem().Kind() == reflect.Uint8 {
			elem.SetBytes(slices.Clone(bytes))
			return nil
		}
	}
	if zreflect.IsBasicKind(kind) {
		ev, err := zreflect.ParseBasicValue(string(bytes), elem.Type())
		if err == nil {
			elem.Set(ev)
		}
		return err
	}

	if kind == reflect.Pointer && zreflect.IsBasicKind(elem.Type().Elem().Kind()) {
		fv := reflect.New(elem.Type().Elem()).Elem()
		ev, err := zreflect.ParseBasicValue(string(bytes), fv.Type())
		if err == nil {
			fv.Set(ev)
			elem.Set(reflect.ValueOf(fv.Addr().Interface()))
		}
		return err
	}
	return fmt.Errorf("type %T not implement TextUnmarshaler", obj)
}
