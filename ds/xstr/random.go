//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-07

package xstr

import (
	"fmt"
	"math/rand/v2"
	"unsafe"
)

const (
	// TableAlphaNum 所有字母和数字的合集
	TableAlphaNum Table = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// TableAlpha 所有字母的合集
	TableAlpha Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type Table string

func (tb Table) RandNChar(n int) string {
	bf := make([]byte, n)
	for i := 0; i < n; i++ {
		bf[i] = tb[rand.IntN(len(tb))]
	}
	return unsafe.String(unsafe.SliceData(bf), n)
}

func (tb Table) RandNByte(n int) []byte {
	bf := make([]byte, n)
	for i := 0; i < n; i++ {
		bf[i] = tb[rand.IntN(len(tb))]
	}
	return bf
}

func (tb Table) RandChar() string {
	n := rand.IntN(len(tb))
	return string(tb[n])
}

func (tb Table) RandByte() byte {
	n := rand.IntN(len(tb))
	return tb[n]
}

// RandNChar 返回一个长度是 n 的字符串
func RandNChar(n int) string {
	return TableAlphaNum.RandNChar(n)
}

func RandNByte(n int) []byte {
	return TableAlphaNum.RandNByte(n)
}

// RandChar 返回一个数字或者字母
func RandChar() string {
	return TableAlphaNum.RandChar()
}

// RandByte 返回一个数字或者字母
func RandByte() byte {
	return TableAlphaNum.RandByte()
}

// RandIdentN 返回一个可用作标志符的字符串
func RandIdentN(n int) string {
	if n < 1 {
		panic(fmt.Sprintf("invalid n %d, should >=1", n))
	}
	bf := make([]byte, n)
	bf[0] = TableAlpha.RandByte()
	for i := 1; i < n; i++ {
		bf[i] = TableAlphaNum.RandByte()
	}
	return unsafe.String(unsafe.SliceData(bf), n)
}
