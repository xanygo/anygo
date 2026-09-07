//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xcodec

import (
	"io"

	"github.com/xanygo/anygo/xcipher"
)

type TransformFunc func([]byte) ([]byte, error)

func (fn TransformFunc) AsWriter(out io.Writer) io.Writer {
	w1 := &tw1{
		onWrite: fn,
		out:     out,
	}
	return w1
}

var _ io.Writer = (*tw1)(nil)

type tw1 struct {
	onWrite TransformFunc
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

type TransformFuncs []func([]byte) ([]byte, error)

func (ts TransformFuncs) Transcoding(data []byte) (result []byte, err error) {
	result = data
	for _, f := range ts {
		result, err = f(result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (ts TransformFuncs) AsWriter(out io.Writer) io.Writer {
	return TransformFunc(ts.Transcoding).AsWriter(out)
}

func EncoderWithTransform(enc Encoder, trans TransformFunc) Encoder {
	return EncodeFunc(func(a any) ([]byte, error) {
		data, err := enc.Encode(a)
		if err != nil {
			return nil, err
		}
		return trans(data)
	})
}

func DecoderWithTransform(dec Decoder, trans TransformFunc) Decoder {
	return DecodeFunc(func(data []byte, obj any) error {
		nd, err := trans(data)
		if err != nil {
			return err
		}
		return dec.Decode(nd, obj)
	})
}

func WithCipher(coder Codec, c xcipher.Cipher) Codec {
	enc := EncoderWithTransform(coder, c.Encrypt)
	dec := DecoderWithTransform(coder, c.Decrypt)
	ct, _ := ContentType(coder)
	return NewCodec(coder.Name(), enc.Encode, dec.Decode, ct)
}
