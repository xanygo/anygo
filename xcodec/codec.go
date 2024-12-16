//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcodec

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

var (
	JSON = NewCodec("json", json.Marshal, json.Unmarshal)

	Raw = NewCodec("raw", rawEncode, rawDecode)
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
)

type EncodeFunc func(any) ([]byte, error)

func (e EncodeFunc) Encode(v any) ([]byte, error) {
	return e(v)
}

type DecodeFunc func([]byte, any) error

func (d DecodeFunc) Decode(v []byte, r any) error {
	return d(v, r)
}

func NewCodec(name string, e EncodeFunc, d DecodeFunc) Codec {
	return &codec{name: name, e: e, d: d}
}

var _ Codec = (*codec)(nil)

type codec struct {
	name string
	e    EncodeFunc
	d    DecodeFunc
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

func rawEncode(obj any) ([]byte, error) {
	switch val := obj.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
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
