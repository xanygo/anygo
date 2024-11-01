//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package znum

import (
	"fmt"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var base62CharMap = map[byte]int{}

func init() {
	for i := 0; i < len(base62Chars); i++ {
		base62CharMap[base62Chars[i]] = i
	}
}

func FormatIntB62(n int64) string {
	if n >= 0 && n <= 9 {
		return string(base62Chars[n])
	}

	var result []byte
	if n < 0 {
		result = append(result, '-')
		n *= -1
	}
	for n > 0 {
		index := n % 62
		result = append(result, base62Chars[index])
		n /= 62
	}
	return string(result)
}

func ParserIntB62(str string) (int64, error) {
	if str == "" {
		return 0, nil
	}
	isNegative := str[0] == '-'
	if isNegative {
		str = str[1:]
	}

	var result int64
	for i := len(str) - 1; i >= 0; i-- {
		char := str[i]
		index, ok := base62CharMap[char]
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
