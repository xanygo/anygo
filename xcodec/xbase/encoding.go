//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xbase

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"unsafe"
)

const (
	Table62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Table36 = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	Base62 = New(Table62)
	Base36 = New(Table36)
)

func New(table string) *Encoding {
	if len(table) < 10 {
		panic("Table length must >= 10, but got " + strconv.Itoa(len(table)))
	}

	index := make(map[byte]int, len(table))
	for i := 0; i < len(table); i++ {
		index[table[i]] = i
	}
	size := int64(len(table))
	return &Encoding{
		table:  table,
		size:   size,
		index:  index,
		biBase: big.NewInt(size),
	}
}

var _ Int64Encoder = (*Encoding)(nil)

type Encoding struct {
	table  string
	size   int64
	index  map[byte]int
	biBase *big.Int
}

func (e *Encoding) Size() int {
	return int(e.size)
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
		index := n % e.size
		result = append(result, e.table[index])
		n /= e.size
	}
	return result
}

var errEmptyStr = errors.New("empty string")

// DecodeInt64String 解析使用 EncodeInt64 编码的字符串
func (e *Encoding) DecodeInt64String(str string) (int64, error) {
	return decodeInt64String(str, e.index, e.size)
}

func (e *Encoding) DecodeInt64Bytes(str []byte) (int64, error) {
	return decodeInt64String(str, e.index, e.size)
}

func decodeInt64String[T string | []byte](str T, tableIndex map[byte]int, size int64) (int64, error) {
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
		result = result*size + int64(index)
	}
	if isNegative {
		return -result, nil
	}
	return result, nil
}

var (
	bi0 = big.NewInt(0)
)

func (e *Encoding) Encode(input []byte) []byte {
	var result []byte
	num := new(big.Int).SetBytes(input)

	for num.Cmp(bi0) > 0 {
		remainder := new(big.Int)
		num.DivMod(num, e.biBase, remainder)
		result = append(result, e.table[remainder.Int64()])
	}
	return result
}

func (e *Encoding) AppendEncode(dst, src []byte) []byte {
	num := new(big.Int).SetBytes(src)

	for num.Cmp(bi0) > 0 {
		remainder := new(big.Int)
		num.DivMod(num, e.biBase, remainder)
		dst = append(dst, e.table[remainder.Int64()])
	}
	return dst
}

func (e *Encoding) EncodeToString(input []byte) string {
	bf := e.Encode(input)
	return unsafe.String(unsafe.SliceData(bf), len(bf))
}

func (e *Encoding) Decode(input []byte) ([]byte, error) {
	return decode(input, e.index, e.biBase)
}

func (e *Encoding) DecodeString(input string) ([]byte, error) {
	return decode(input, e.index, e.biBase)
}

func decode[T string | []byte](input T, tableIndex map[byte]int, base *big.Int) ([]byte, error) {
	result := big.NewInt(0)

	for i := len(input) - 1; i >= 0; i-- {
		char := input[i]
		idx, ok := tableIndex[char]
		if !ok {
			return nil, fmt.Errorf("invalid character: %c", char)
		}
		result.Mul(result, base)
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
