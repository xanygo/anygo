//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xenc

import (
	"io"
)

type TextTransformer interface {
	TextEncoder
	TextDecoder
}

type TextEncoder interface {
	Encode(input []byte) ([]byte, error)
}

type TextDecoder interface {
	Decode(input []byte) ([]byte, error)
}

type TextTransformFunc func([]byte) ([]byte, error)

func (fn TextTransformFunc) AsWriter(out io.Writer) io.Writer {
	w1 := &tw1{
		onWrite: fn,
		out:     out,
	}
	return w1
}

func (fn TextTransformFunc) Encode(input []byte) ([]byte, error) {
	return fn(input)
}

func (fn TextTransformFunc) Decode(input []byte) ([]byte, error) {
	return fn(input)
}

var _ io.Writer = (*tw1)(nil)

type tw1 struct {
	onWrite TextTransformFunc
	out     io.Writer
}

func (w *tw1) Write(p []byte) (n int, err error) {
	ep, err := w.onWrite(p)
	if err != nil {
		return 0, err
	}
	_, err1 := w.out.Write(ep)
	if err1 != nil {
		return 0, err1
	}
	return len(p), nil
}

type TextTransformFuncs []func([]byte) ([]byte, error)

func (ts TextTransformFuncs) transcoding(data []byte) (result []byte, err error) {
	result = data
	for _, f := range ts {
		result, err = f(result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (ts TextTransformFuncs) AsWriter(out io.Writer) io.Writer {
	return TextTransformFunc(ts.transcoding).AsWriter(out)
}

func (ts TextTransformFuncs) Encode(input []byte) ([]byte, error) {
	return ts.transcoding(input)
}

func (ts TextTransformFuncs) Decode(input []byte) ([]byte, error) {
	return ts.transcoding(input)
}

func MarshalerWithTransform(enc Marshaler, trans TextTransformFunc) Marshaler {
	return MarshalFunc(func(a any) ([]byte, error) {
		data, err := enc.Marshal(a)
		if err != nil {
			return nil, err
		}
		return trans(data)
	})
}

func UnmarshalerWithTransform(dec Unmarshaler, trans TextTransformFunc) Unmarshaler {
	return UnmarshalFunc(func(data []byte, obj any) error {
		nd, err := trans(data)
		if err != nil {
			return err
		}
		return dec.Unmarshal(nd, obj)
	})
}

type TextMapperFunc func([]byte) []byte

func (fn TextMapperFunc) Encode(input []byte) ([]byte, error) {
	return fn(input), nil
}

func (fn TextMapperFunc) Encrypt(input []byte) ([]byte, error) {
	return fn(input), nil
}

func (fn TextMapperFunc) Decode(input []byte) ([]byte, error) {
	return fn(input), nil
}

func (fn TextMapperFunc) Decrypt(input []byte) ([]byte, error) {
	return fn(input), nil
}
