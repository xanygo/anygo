//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-17

package xurl

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func String(query url.Values, name string) (string, error) {
	if len(query) == 0 {
		return "", errors.New("empty values")
	}
	vs := query[name]
	if len(vs) == 0 {
		return "", fmt.Errorf("no value for %q", name)
	}
	return vs[0], nil
}

func StringDef(query url.Values, name string, def string) string {
	str, err := String(query, name)
	if err == nil {
		return str
	}
	return def
}

func Strings(query url.Values, name string, sep string) []string {
	str, err := String(query, name)
	if err != nil {
		return nil
	}
	arr := strings.Split(str, sep)
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		v := strings.TrimSpace(v)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

func Int(query url.Values, name string) (int, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func IntDef(query url.Values, name string, def int) int {
	num, err := Int(query, name)
	if err == nil {
		return num
	}
	return def
}

func Ints(query url.Values, name string, sep string) ([]int, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]int, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.Atoi(v)
		if err == nil {
			return nil, err
		}
		result = append(result, num)
	}
	return result, nil
}

func Int8(query url.Values, name string) (int8, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseInt(str, 10, 8)
	return int8(num), err
}

func Int8Def(query url.Values, name string, def int8) int8 {
	num, err := Int8(query, name)
	if err == nil {
		return num
	}
	return def
}

func Int8s(query url.Values, name string, sep string) ([]int8, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]int8, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseInt(v, 10, 8)
		if err == nil {
			return nil, err
		}
		result = append(result, int8(num))
	}
	return result, nil
}

func Int16(query url.Values, name string) (int16, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseInt(str, 10, 16)
	return int16(num), err
}

func Int16Def(query url.Values, name string, def int16) int16 {
	num, err := Int16(query, name)
	if err == nil {
		return num
	}
	return def
}

func Int16s(query url.Values, name string, sep string) ([]int16, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]int16, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseInt(v, 10, 16)
		if err == nil {
			return nil, err
		}
		result = append(result, int16(num))
	}
	return result, nil
}

func Int32(query url.Values, name string) (int32, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseInt(str, 10, 32)
	return int32(num), err
}

func Int32Def(query url.Values, name string, def int32) int32 {
	num, err := Int32(query, name)
	if err == nil {
		return num
	}
	return def
}

func Int32s(query url.Values, name string, sep string) ([]int32, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]int32, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			return nil, err
		}
		result = append(result, int32(num))
	}
	return result, nil
}

func Int64(query url.Values, name string) (int64, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(str, 10, 64)
}

func Int64Def(query url.Values, name string, def int64) int64 {
	num, err := Int64(query, name)
	if err == nil {
		return num
	}
	return def
}

func Int64s(query url.Values, name string, sep string) ([]int64, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]int64, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return nil, err
		}
		result = append(result, num)
	}
	return result, nil
}

func Uint(query url.Values, name string) (uint, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseUint(str, 10, 0)
	return uint(num), err
}

func UintDef(query url.Values, name string, def uint) uint {
	num, err := Uint(query, name)
	if err == nil {
		return num
	}
	return def
}

func Uints(query url.Values, name string, sep string) ([]uint, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]uint, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseUint(v, 10, 0)
		if err == nil {
			return nil, err
		}
		result = append(result, uint(num))
	}
	return result, nil
}

func Uint8(query url.Values, name string) (uint8, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseUint(str, 10, 8)
	return uint8(num), err
}

func Uint8Def(query url.Values, name string, def uint8) uint8 {
	num, err := Uint8(query, name)
	if err == nil {
		return num
	}
	return def
}

func Uint8s(query url.Values, name string, sep string) ([]uint8, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]uint8, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseUint(v, 10, 8)
		if err == nil {
			return nil, err
		}
		result = append(result, uint8(num))
	}
	return result, nil
}

func Uint16(query url.Values, name string) (uint16, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseUint(str, 10, 16)
	return uint16(num), err
}

func Uint16Def(query url.Values, name string, def uint16) uint16 {
	num, err := Uint16(query, name)
	if err == nil {
		return num
	}
	return def
}

func Uint16s(query url.Values, name string, sep string) ([]uint16, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]uint16, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseUint(v, 10, 16)
		if err == nil {
			return nil, err
		}
		result = append(result, uint16(num))
	}
	return result, nil
}

func Uint32(query url.Values, name string) (uint32, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseUint(str, 10, 32)
	return uint32(num), err
}

func Uint32Def(query url.Values, name string, def uint32) uint32 {
	num, err := Uint32(query, name)
	if err == nil {
		return num
	}
	return def
}

func Uint32s(query url.Values, name string, sep string) ([]uint32, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]uint32, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			return nil, err
		}
		result = append(result, uint32(num))
	}
	return result, nil
}

func Uint64(query url.Values, name string) (uint64, error) {
	str, err := String(query, name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(str, 10, 64)
}

func Uint64Def(query url.Values, name string, def uint64) uint64 {
	num, err := Uint64(query, name)
	if err == nil {
		return num
	}
	return def
}

func Uint64s(query url.Values, name string, sep string) ([]uint64, error) {
	ss := Strings(query, name, sep)
	if len(ss) == 0 {
		return nil, nil
	}
	result := make([]uint64, 0, len(ss))
	for _, v := range ss {
		num, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			return nil, err
		}
		result = append(result, num)
	}
	return result, nil
}

// Page 读取 query 中的页码参数，总是返回 >=1 的值
func Page(query url.Values, name string) int {
	return ParserPage(query.Get(name))
}

// ParserPage 解析 page 参数，总是返回 >=1 的值
func ParserPage(str string) int {
	num, _ := strconv.Atoi(str)
	if num < 1 {
		return 1
	}
	return num
}
