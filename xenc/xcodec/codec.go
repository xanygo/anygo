//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcodec

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/xanygo/anygo/xenc"
)

type Codec = xenc.Codec

type Marshaler = xenc.Marshaler

type Unmarshaler = xenc.Unmarshaler

type UnmarshalExtra = xenc.UnmarshalExtra

// MarshalToString 使用 Marshaler 将 obj 编码为 字符串
func MarshalToString(enc Marshaler, obj any) (string, error) {
	bf, err := enc.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("encode error %w, data=%#v", err, obj)
	}
	return unsafe.String(unsafe.SliceData(bf), len(bf)), nil
}

// UnmarshalFromString 使用 Decoder 将字符串 解码并赋值给 obj，若 obj 本身是字符串类型，则直接赋值
func UnmarshalFromString(dec Unmarshaler, str string, obj any) error {
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	return Unmarshal(dec, bf, obj)
}

func Unmarshal(decoder Unmarshaler, content []byte, obj any) error {
	err := decoder.Unmarshal(content, obj)
	if err != nil {
		return fmt.Errorf("xcodec.Decoder %q, %d bytes, %w", xenc.Name(decoder), len(content), err)
	}
	return doDecodeExtra(decoder, content, obj)
}

// doDecodeExtra 若obj 实现了 ParseExtra，则将其为定义字段解析到 Extra（具体字段名由 ParseExtra 接口返回） 里去
func doDecodeExtra(decoder Unmarshaler, content []byte, obj any) error {
	et, ok := obj.(UnmarshalExtra)
	if !ok {
		return nil
	}
	name := et.NeedDecodeExtra()
	if name == "" {
		return nil
	}
	rt := reflect.TypeOf(obj).Elem()
	fieldType, ok := rt.FieldByName(name)
	if !ok {
		return fmt.Errorf("filed %q not exixts", name)
	}
	if !isMapStringAny(fieldType.Type) {
		return fmt.Errorf("filed %q is not map[string]any", name)
	}

	rv := reflect.ValueOf(obj).Elem()
	field := rv.FieldByName(name)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("filed %q is not settable", name)
	}

	data := map[string]any{}
	if err := decoder.Unmarshal(content, &data); err != nil {
		return err
	}
	names := make(map[string]bool, rt.NumField())
	for field := range rt.Fields() {
		// 在比较字段名时，全部转换为小写。以避免如json、yaml等解析时，tag 定义的名字和字段名不一直的情况
		fn := strings.ToLower(field.Name)
		names[fn] = true
	}

	if field.IsNil() { // 如果没初始化，先初始化
		field.Set(reflect.MakeMap(field.Type()))
	}
	for k, v := range data {
		if names[strings.ToLower(k)] {
			continue
		}
		key := reflect.ValueOf(k)
		val := reflect.ValueOf(v)
		field.SetMapIndex(key, val)
	}
	return nil
}

func isMapStringAny(t reflect.Type) bool {
	return t.Kind() == reflect.Map &&
		t.Key().Kind() == reflect.String &&
		t.Elem().Kind() == reflect.Interface
}

func Marshal(enc Marshaler, obj any) ([]byte, error) {
	return enc.Marshal(obj)
}
