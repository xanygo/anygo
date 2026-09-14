//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xcodec

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/xanygo/anygo/internal/zreflect"
)

var Text = TextCodec{}

var _ Codec = (*TextCodec)(nil)

type TextCodec struct{}

func (t TextCodec) Name() string {
	return "text"
}

func (t TextCodec) ContentType() string {
	return "text/plain"
}

func (t TextCodec) TextMarshal(obj any) (bytes []byte, err error, match bool) {
	if mt, ok := obj.(encoding.TextMarshaler); ok {
		bytes, err = mt.MarshalText()
		return bytes, err, true
	}

	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Pointer {
		p := reflect.New(rv.Type())
		p.Elem().Set(rv)
		if mt, ok := p.Interface().(encoding.TextMarshaler); ok {
			bytes, err = mt.MarshalText()
			return bytes, err, true
		}
	}
	return nil, nil, false
}

func (t TextCodec) Marshal(obj any) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("nil value")
	}

	bytes, err, match := t.TextMarshal(obj)
	if match {
		return bytes, err
	}

	if str, ok := zreflect.BaseTypeToString(obj); ok {
		return []byte(str), nil
	}
	if b, ok := zreflect.BytesValue(obj); ok {
		return b, nil
	}
	return nil, fmt.Errorf("type %T cannot Marshal by TextCodec", obj)
}

// Unmarshal 将bytes 解析到 obj，具体规则如下：
//
//  1. obj 实现了 TextUnmarshaler ，则优先使用
//  2. 若 obj 是  string 或者 []byte 类型，则直接赋值
//  3. 若 obj 是基础类型，如 number、bool 类型，则尝试解析赋值
//  4. 返回错误
func (t TextCodec) Unmarshal(bytes []byte, obj any) error {
	return t.UnmarshalCustom(bytes, obj, true)
}

// 取个合适的名字
func (t TextCodec) UnmarshalCustom(bytes []byte, obj any, custom bool) error {
	rv := reflect.ValueOf(obj)
	for rv.Kind() == reflect.Pointer {
		if custom && rv.CanAddr() {
			if u, ok := rv.Interface().(encoding.TextUnmarshaler); ok {
				if rv.Kind() == reflect.Pointer && rv.IsNil() {
					if !rv.CanSet() {
						return fmt.Errorf("cannot initialize %T", obj)
					}
					rv.Set(reflect.New(rv.Type().Elem()))
					u = rv.Interface().(encoding.TextUnmarshaler)
				}
				return u.UnmarshalText(bytes)
			}
		}
		if rv.IsNil() {
			if !rv.CanSet() {
				return fmt.Errorf("cannot initialize nil pointer %T", obj)
			}
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	if custom && rv.CanAddr() && rv.Addr().CanInterface() {
		if u, ok := rv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.UnmarshalText(bytes)
		}
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(string(bytes))
		return nil
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			rv.SetBytes(slices.Clone(bytes))
			return nil
		}
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			if len(bytes) > rv.Len() {
				return fmt.Errorf("cannot decode into %T with %d bytes", obj, len(bytes))
			}
			reflect.Copy(rv, reflect.ValueOf(slices.Clone(bytes)))
			return nil
		}
	default:
		// pass
	}

	if rv.CanSet() && zreflect.IsBasicKind(rv.Kind()) {
		ev, err := zreflect.ParseBasicValue(string(bytes), rv.Type())
		if err == nil {
			rv.Set(ev)
		}
		return err
	}

	return fmt.Errorf("type %T cannot Unmarshal by TextCodec", obj)
}
