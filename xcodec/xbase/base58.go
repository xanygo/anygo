//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-08-18

package xbase

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"unsafe"
)

// Table58 常见于比特币和区块链领域，用来把二进制数据（比如哈希值、公钥、地址等）转化为更便于人类使用的字符串
// 去掉了 0（数字零）、O（大写字母 O）、I（大写字母 i）、l（小写字母 L），避免视觉混淆
// 不包含 + 和 /，更适合 URL 或命令行
const Table58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var Base58 = &Base58Codec{}

type Base58Codec struct{}

func (b *Base58Codec) Encode(input []byte) []byte {
	var result []byte

	// 转换成大整数
	x := new(big.Int).SetBytes(input)

	// 循环除以 58，取余数
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, Table58[mod.Int64()])
	}

	// 处理前导 0（即字节开头的 0x00 转换为 '1'）
	for _, b := range input {
		if b == 0x00 {
			result = append(result, Table58[0])
		} else {
			break
		}
	}

	// Base58 编码是反序的，要翻转
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func (b *Base58Codec) EncodeToString(input []byte) string {
	bf := b.Encode(input)
	return unsafe.String(unsafe.SliceData(bf), len(bf))
}

func (b *Base58Codec) AppendEncode(dst, src []byte) []byte {
	encoded := b.Encode(src)
	return append(dst, encoded...)
}

func (b *Base58Codec) Decode(input []byte) ([]byte, error) {
	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, r := range input {
		charIndex := strings.IndexByte(Table58, r)
		if charIndex < 0 {
			return nil, fmt.Errorf("invalid Base58 character: %c", r)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	// 转回字节
	decoded := result.Bytes()

	// 处理前导 '1'（代表 0x00）
	nLeadingZeros := 0
	for _, r := range input {
		if r == Table58[0] {
			nLeadingZeros++
		} else {
			break
		}
	}

	return append(bytes.Repeat([]byte{0x00}, nLeadingZeros), decoded...), nil
}

func (b *Base58Codec) DecodeString(input string) ([]byte, error) {
	bf := unsafe.Slice(unsafe.StringData(input), len(input))
	return b.Decode(bf)
}

func (b *Base58Codec) AppendDecode(dst, src []byte) ([]byte, error) {
	result, err := b.Decode(src)
	if err != nil {
		return nil, err
	}
	return append(dst, result...), nil
}
