//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-21

package xhash

import (
	"crypto/md5"
	"encoding/hex"
	"unsafe"
)

func Md5(s string) string {
	bf := unsafe.Slice(unsafe.StringData(s), len(s))
	h := md5.Sum(bf)
	return hex.EncodeToString(h[:])
}

func ByteMd5(b []byte) string {
	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}
