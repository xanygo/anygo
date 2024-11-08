//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xstr

import (
	"math/rand/v2"
	"unsafe"
)

// TableAlphaNum 所有字母和数字的合集
const TableAlphaNum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func RandomN(n int) string {
	bf := make([]byte, n)
	for i := 0; i < n; i++ {
		bf[i] = TableAlphaNum[rand.IntN(len(TableAlphaNum))]
	}
	return unsafe.String(&bf[0], n)
}
