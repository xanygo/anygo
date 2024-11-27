//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xcodec

import (
	"github.com/xanygo/anygo/xcodec/xbase"
)

var _ xbase.Int64Encoder = (*Int64Cipher)(nil)

// Int64Cipher 将 int64 加密为字符串的算法
type Int64Cipher struct {
	// Cipher 必填 加密套件
	Cipher Cipher

	// Int64Encoder 可选编码器，默认为 xbase.Base62
	Int64Encoder *xbase.Encoding
}

func (n *Int64Cipher) EncodeInt64(num int64) string {
	str, _ := n.Encode(num)
	return str
}

func (n *Int64Cipher) EncodeInt64Byte(num int64) []byte {
	enc := n.getEncoder()
	bf := enc.EncodeInt64Byte(num)
	out, err := n.Cipher.Encrypt(bf)
	if err != nil {
		return nil
	}
	return enc.Encode(out)
}

func (n *Int64Cipher) DecodeInt64String(str string) (int64, error) {
	return n.Decode(str)
}

func (n *Int64Cipher) DecodeInt64Bytes(str []byte) (int64, error) {
	enc := n.getEncoder()
	bf, err := enc.Decode(str)
	if err != nil {
		return 0, err
	}
	out, err := n.Cipher.Decrypt(bf)
	if err != nil {
		return 0, err
	}
	return enc.DecodeInt64Bytes(out)
}

func (n *Int64Cipher) Encode(num int64) (string, error) {
	enc := n.getEncoder()
	bf := enc.EncodeInt64Byte(num)
	out, err := n.Cipher.Encrypt(bf)
	if err != nil {
		return "", err
	}
	return enc.EncodeToString(out), nil
}

func (n *Int64Cipher) Decode(str string) (int64, error) {
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

func (n *Int64Cipher) getEncoder() *xbase.Encoding {
	if n.Int64Encoder == nil {
		return xbase.Base62
	}
	return n.Int64Encoder
}
