//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcodec

import (
	"encoding/json"
	"fmt"
	"net/url"
	"unsafe"
)

var (
	JSON = NewCodec("json", json.Marshal, json.Unmarshal, "application/json")

	Raw = NewCodec("raw", rawEncode, rawDecode, "application/octet-stream")

	Form = &FormCodec{}
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

func EncodeToString(enc Encoder, obj any) (string, error) {
	bf, err := enc.Encode(obj)
	if err != nil {
		return "", err
	}
	return unsafe.String(unsafe.SliceData(bf), len(bf)), nil
}

func DecodeFromString(dec Decoder, str string, obj any) error {
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	return dec.Decode(bf, obj)
}
