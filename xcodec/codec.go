//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcodec

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"unsafe"
)

var (
	JSON = NewCodec("json", json.Marshal, json.Unmarshal, "application/json")

	Raw = NewCodec("raw", rawEncode, rawDecode, "application/octet-stream")

	Form = &FormCodec{}

	CSV = CSVCodec{}

	Text = TextCodec{}
)

type (
	Codec interface {
		Name() string
		Encoder
		Decoder
	}

	Encoder interface {
		Encode(any) ([]byte, error)
	}

	Decoder interface {
		Decode([]byte, any) error
	}

	HasContentType interface {
		ContentType() string
	}
)

type EncodeFunc func(any) ([]byte, error)

func (e EncodeFunc) Encode(v any) ([]byte, error) {
	return e(v)
}

type DecodeFunc func([]byte, any) error

func (d DecodeFunc) Decode(v []byte, r any) error {
	return d(v, r)
}

func NewCodec(name string, e EncodeFunc, d DecodeFunc, ct string) Codec {
	return &codec{name: name, e: e, d: d, ct: ct}
}

var _ Codec = (*codec)(nil)
var _ HasContentType = (*codec)(nil)

type codec struct {
	name string
	e    EncodeFunc
	d    DecodeFunc
	ct   string
}

func (c *codec) Encode(a any) ([]byte, error) {
	return c.e(a)
}

func (c *codec) Decode(bf []byte, a any) error {
	return c.d(bf, a)
}

func (c *codec) Name() string {
	return c.name
}

func (c *codec) ContentType() string {
	return c.ct
}

func rawEncode(obj any) ([]byte, error) {
	switch val := obj.(type) {
	case []byte:
		return val, nil
	case string:
		return unsafe.Slice(unsafe.StringData(val), len(val)), nil
	default:
		return nil, fmt.Errorf("not support type %T for rawEncode", obj)
	}
}

func rawDecode(data []byte, obj any) error {
	switch val := obj.(type) {
	case *[]byte:
		*val = data
		return nil
	case *string:
		*val = string(data)
		return nil
	default:
		return fmt.Errorf("not support type %T for rawDecode", obj)
	}
}

func JSONString(obj any) string {
	bf, err := json.Marshal(obj)
	if err != nil {
		return err.Error()
	}
	return unsafe.String(unsafe.SliceData(bf), len(bf))
}

var _ Codec = (*FormCodec)(nil)
var _ HasContentType = (*FormCodec)(nil)

type FormCodec struct {
}

func (f FormCodec) Name() string {
	return "Form"
}

func (f FormCodec) ContentType() string {
	return "application/x-www-form-urlencoded"
}

func (f FormCodec) Encode(a any) ([]byte, error) {
	switch vv := a.(type) {
	case url.Values:
		str := vv.Encode()
		return unsafe.Slice(unsafe.StringData(str), len(str)), nil
	case map[string]string:
		uv := make(url.Values, len(vv))
		for k, v := range vv {
			uv.Set(k, v)
		}
		str := uv.Encode()
		return unsafe.Slice(unsafe.StringData(str), len(str)), nil
	default:
		return nil, fmt.Errorf("not support type %T for FormEncode", a)
	}
}

func (f FormCodec) Decode(bf []byte, a any) error {
	if len(bf) == 0 {
		return nil
	}
	values, err := url.ParseQuery(string(bf))
	if err != nil {
		return err
	}

	switch vv := a.(type) {
	case *url.Values:
		*vv = values
		return nil
	case *map[string]string:
		m := make(map[string]string, len(values))
		for k, v := range values {
			if len(v) > 0 {
				m[k] = v[0] // 只取第一个
			}
		}
		*vv = m
		return nil
	default:
		return fmt.Errorf("not support type %T for FormDecode", a)
	}
}

// EncodeToString 使用 Encoder 将 obj 编码为 字符串，若 obj 本身就是字符串，则直接返回
func EncodeToString(enc Encoder, obj any) (string, error) {
	rv := reflect.ValueOf(obj)
	for {
		if rv.Kind() == reflect.String {
			return rv.String(), nil
		} else if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return "", nil
			}
			rv = rv.Elem()
		} else {
			break
		}
	}

	// 由于无法区分开 []byte 和 []uint8 ，所以不对 []byte 做特殊处理

	bf, err := enc.Encode(obj)
	if err != nil {
		return "", err
	}
	return unsafe.String(unsafe.SliceData(bf), len(bf)), nil
}

// DecodeFromString 使用 Decoder 将字符串 解码并赋值给 obj，若 obj 本身是字符串类型，则直接赋值
func DecodeFromString(dec Decoder, str string, obj any) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("obj must be a non-nil pointer, got %T", obj)
	}
	elem := rv.Elem()
	if elem.Kind() == reflect.String {
		elem.SetString(str)
		return nil
	}

	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	return Decode(dec, bf, obj)
}

// DecodeExtra 当被解析的对象，实现了此接口的时候，并且 NeedDecodeExtra 返回了有效的字段名，
// 则会将未在 struct 中定义的字段，全部解析到指定的字段里。
type DecodeExtra interface {
	// NeedDecodeExtra 存储未定义字段的字段名，返回非空为有效。
	// 并且返回的名字必须在 struct 中存在，而且必须是 map[string]any 类型
	NeedDecodeExtra() string
}

func Decode(decoder Decoder, content []byte, obj any) error {
	err := decoder.Decode(content, obj)
	if err != nil {
		return err
	}
	return doDecodeExtra(decoder, content, obj)
}

// doDecodeExtra 若obj 实现了 ParseExtra，则将其为定义字段解析到 Extra（具体字段名由 ParseExtra 接口返回） 里去
func doDecodeExtra(decoder Decoder, content []byte, obj any) error {
	et, ok := obj.(DecodeExtra)
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
	if err := decoder.Decode(content, &data); err != nil {
		return err
	}
	names := make(map[string]bool, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		// 在比较字段名时，全部转换为小写。以避免如json、yaml等解析时，tag 定义的名字和字段名不一直的情况
		fn := strings.ToLower(rt.Field(i).Name)
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

var errNoCt = errors.New("invalid codec: not xcodec.HasContentType")

func ContentType(c Encoder) (string, error) {
	if hct, ok := c.(HasContentType); ok {
		return hct.ContentType(), nil
	}
	return "", errNoCt
}
