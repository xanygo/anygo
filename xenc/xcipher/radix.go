//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xcipher

import (
	"fmt"
	"reflect"

	"github.com/xanygo/anygo/internal/zbase"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xenc"
)

var _ xenc.RadixCodec = (*Int64)(nil)

// Int64 将 int64 加密为字符串的算法
type Int64 struct {
	// Cipher 必填 加密套件
	Cipher xenc.Cipher

	// Encoder 可选编码器，默认为 xencoding.Base62
	Encoder *zbase.Encoding
}

func (n *Int64) Name() string {
	return "Int64Cipher"
}

func (n *Int64) EncodeInt(num int64) (string, error) {
	enc := n.getEncoder()
	bf := enc.EncodeInt64Byte(num)
	out, err := n.Cipher.Encrypt(bf)
	if err != nil {
		return "", err
	}
	return enc.EncodeToString(out), nil
}

func (n *Int64) DecodeInt(str string) (int64, error) {
	enc := n.getEncoder()
	bf, err := enc.DecodeString(str)
	if err != nil {
		return 0, err
	}
	out, err := n.Cipher.Decrypt(bf)
	if err != nil {
		return 0, err
	}
	return enc.DecodeInt64Bytes(out)
}

var _ xenc.Codec = (*Int64)(nil)

func (n *Int64) Marshal(num any) ([]byte, error) {
	nv, ok := zreflect.BaseTypeToInt64(num)
	if !ok {
		return nil, fmt.Errorf("cannot Encode %#v", num)
	}
	enc := n.getEncoder()
	bf := enc.EncodeInt64Byte(nv)
	out, err := n.Cipher.Encrypt(bf)
	if err != nil {
		return nil, err
	}
	return enc.Encode(out), nil
}

func (n *Int64) Unmarshal(str []byte, obj any) error {
	rv := reflect.ValueOf(obj)
	if !rv.IsValid() {
		return fmt.Errorf("invalid value %v", obj)
	}
	enc := n.getEncoder()
	bf, err := enc.Decode(str)
	if err != nil {
		return err
	}
	out, err := n.Cipher.Decrypt(bf)
	if err != nil {
		return err
	}
	num, err := enc.DecodeInt64Bytes(out)
	if err != nil {
		return err
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.OverflowInt(num) {
			return fmt.Errorf("%d overflows %s", num, rv.Type())
		}
		rv.SetInt(num)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if num < 0 || rv.OverflowUint(uint64(num)) {
			return fmt.Errorf("%d overflows %s", num, rv.Type())
		}
		rv.SetUint(uint64(num))
	}
	return nil
}

func (n *Int64) getEncoder() *zbase.Encoding {
	if n.Encoder == nil {
		return zbase.Base62
	}
	return n.Encoder
}
