//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xbase62

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"unsafe"
)

const Table = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func New(table string) *Encoding {
	if len(table) != 62 {
		panic("Table length must be 62, but got " + strconv.Itoa(len(table)))
	}
	index := make(map[byte]int, 62)
	for i := 0; i < len(table); i++ {
		index[table[i]] = i
	}
	return &Encoding{
		table: table,
		index: index,
	}
}

type Encoding struct {
	table string
	index map[byte]int
}

// EncodeInt64 将 int64 编码为字符串
func (e *Encoding) EncodeInt64(n int64) string {
	bf := e.EncodeInt64Byte(n)
	return unsafe.String(unsafe.SliceData(bf), len(bf))
}

func (e *Encoding) EncodeInt64Byte(n int64) []byte {
	if n >= 0 && n <= 9 {
		return []byte{e.table[n]}
	}

	var result []byte
	if n < 0 {
		result = append(result, '-')
		n *= -1
	}
	for n > 0 {
		index := n % 62
		result = append(result, e.table[index])
		n /= 62
	}
	return result
}

var errEmptyStr = errors.New("empty string")

// DecodeInt64String 解析使用 EncodeInt64 编码的字符串
func (e *Encoding) DecodeInt64String(str string) (int64, error) {
	return decodeInt64String(str, e.index)
}

func (e *Encoding) DecodeInt64Bytes(str []byte) (int64, error) {
	return decodeInt64String(str, e.index)
}

func decodeInt64String[T string | []byte](str T, tableIndex map[byte]int) (int64, error) {
	if len(str) == 0 {
		return 0, errEmptyStr
	}
	isNegative := str[0] == '-'
	if isNegative {
		str = str[1:]
	}

	var result int64
	for i := len(str) - 1; i >= 0; i-- {
		char := str[i]
		index, ok := tableIndex[char]
		if !ok {
			return 0, fmt.Errorf("invalid character %c", char)
		}
		result = result*62 + int64(index)
	}
	if isNegative {
		return -result, nil
	}
	return result, nil
}

var (
	bi0  = big.NewInt(0)
	bi62 = big.NewInt(62)
)

func (e *Encoding) Encode(input []byte) []byte {
	var result []byte
	num := new(big.Int).SetBytes(input)

	for num.Cmp(bi0) > 0 {
		remainder := new(big.Int)
		num.DivMod(num, bi62, remainder)
		result = append(result, e.table[remainder.Int64()])
	}
	return result
}

func (e *Encoding) AppendEncode(dst, src []byte) []byte {
	num := new(big.Int).SetBytes(src)

	for num.Cmp(bi0) > 0 {
		remainder := new(big.Int)
		num.DivMod(num, bi62, remainder)
		dst = append(dst, e.table[remainder.Int64()])
	}
	return dst
}

func (e *Encoding) EncodeToString(input []byte) string {
	bf := e.Encode(input)
	return unsafe.String(unsafe.SliceData(bf), len(bf))
}

func (e *Encoding) Decode(input []byte) ([]byte, error) {
	return decode(input, e.index)
}

func (e *Encoding) DecodeString(input string) ([]byte, error) {
	return decode(input, e.index)
}

func decode[T string | []byte](input T, tableIndex map[byte]int) ([]byte, error) {
	result := big.NewInt(0)

	for i := len(input) - 1; i >= 0; i-- {
		char := input[i]
		idx, ok := tableIndex[char]
		if !ok {
			return nil, fmt.Errorf("invalid character: %c", char)
		}
		result.Mul(result, bi62)
		result.Add(result, big.NewInt(int64(idx)))
	}

	return result.Bytes(), nil
}

func (e *Encoding) AppendDecode(dst, src []byte) ([]byte, error) {
	result, err := e.Decode(src)
	if err != nil {
		return nil, err
	}
	return append(dst, result...), nil
}

var Default = New(Table)

func Random1() string {
	n := rand.IntN(len(Table))
	return string(Table[n])
}

func EncodeInt64(n int64) string {
	return Default.EncodeInt64(n)
}

func EncodeInt64Byte(n int64) []byte {
	return Default.EncodeInt64Byte(n)
}

func DecodeInt64String(str string) (int64, error) {
	return Default.DecodeInt64String(str)
}

func DecodeInt64Bytes(str []byte) (int64, error) {
	return Default.DecodeInt64Bytes(str)
}

func Encode(input []byte) []byte {
	return Default.Encode(input)
}

func Decode(input []byte) ([]byte, error) {
	return Default.Decode(input)
}

func EncodeToString(input []byte) string {
	return Default.EncodeToString(input)
}

func DecodeString(input string) ([]byte, error) {
	return Default.DecodeString(input)
}
