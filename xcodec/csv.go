//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xcodec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/internal/zreflect"
)

var _ Codec = (*CSVCodec)(nil)
var _ HasContentType = (*CSVCodec)(nil)

type CSVCodec struct {
}

func (c CSVCodec) ContentType() string {
	return "text/csv"
}

func (c CSVCodec) Name() string {
	return "csv"
}

func (c CSVCodec) Encode(a any) ([]byte, error) {
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expect slice but got %T", a)
	}

	n := v.Len()
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		val := v.Index(i).Interface()
		if str, ok := zreflect.BaseTypeToString(val); ok {
			parts[i] = str
		} else {
			return nil, fmt.Errorf("invalid value %#v", val)
		}
	}

	return []byte(strings.Join(parts, ",")), nil
}

func (c CSVCodec) Decode(b []byte, a any) error {
	if len(b) == 0 {
		return nil // 空值表示空 slice
	}

	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("expect pointer but got %T", a)
	}

	sliceType := rv.Elem().Type()
	if sliceType.Kind() != reflect.Slice {
		return fmt.Errorf("expect pointer to slice but got %T", a)
	}

	elemType := sliceType.Elem()
	parts := strings.Split(string(b), ",")
	slice := reflect.MakeSlice(sliceType, len(parts), len(parts))

	for i, s := range parts {
		s = strings.TrimSpace(s)
		val, err := zreflect.ParseBasicValue(s, elemType)
		if err != nil {
			return fmt.Errorf("csv.Decode(%q): %w", s, err)
		}
		slice.Index(i).Set(val)
	}

	rv.Elem().Set(slice)
	return nil
}
