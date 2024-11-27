//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xstr

import (
	"math/rand/v2"
	"unsafe"
)

// TableAlphaNum 所有字母和数字的合集
const TableAlphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandNChar 返回一个长度是 n 的字符串
func RandNChar(n int) string {
	bf := make([]byte, n)
	for i := 0; i < n; i++ {
		bf[i] = TableAlphaNum[rand.IntN(len(TableAlphaNum))]
	}
	return unsafe.String(unsafe.SliceData(bf), n)
}
func RandNByte(n int) []byte {
	bf := make([]byte, n)
	for i := 0; i < n; i++ {
		bf[i] = TableAlphaNum[rand.IntN(len(TableAlphaNum))]
	}
	return bf
}

// RandChar 返回一个数字或者字母
func RandChar() string {
	n := rand.IntN(len(TableAlphaNum))
	return string(TableAlphaNum[n])
}

// RandByte 返回一个数字或者字母
func RandByte() byte {
	n := rand.IntN(len(TableAlphaNum))
	return TableAlphaNum[n]
}
