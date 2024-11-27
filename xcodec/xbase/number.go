//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-27

package xbase

type Int64Encoder interface {
	EncodeInt64(n int64) string
	EncodeInt64Byte(n int64) []byte

	DecodeInt64String(str string) (int64, error)
	DecodeInt64Bytes(str []byte) (int64, error)
}
