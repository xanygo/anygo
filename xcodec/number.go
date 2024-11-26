//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-26

package xcodec

import (
	"github.com/xanygo/anygo/xcodec/xbase62"
)

// Int64Cipher 将 int64 加密为字符串的算法
type Int64Cipher struct {
	Cipher Cipher // 必填 加密套件
}

func (n *Int64Cipher) Encode(num int64) (string, error) {
	bf := xbase62.EncodeInt64Byte(num)
	out, err := n.Cipher.Encrypt(bf)
	if err != nil {
		return "", err
	}
	return xbase62.EncodeToString(out), nil
}

func (n *Int64Cipher) Decode(str string) (int64, error) {
	bf, err := xbase62.DecodeString(str)
	if err != nil {
		return 0, err
	}
	out, err := n.Cipher.Decrypt(bf)
	if err != nil {
		return 0, err
	}
	return xbase62.DecodeInt64Bytes(out)
}
