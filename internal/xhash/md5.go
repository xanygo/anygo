//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-21

package xhash

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5[T ~string | ~[]byte](s T) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
