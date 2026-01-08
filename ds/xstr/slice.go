//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xstr

import (
	"strconv"
	"strings"
)

// ToSliceFunc 将字符串使用 sep 拆分，并使用 fn 对每个子串解析，返回解析后的内容
//
//	使用 sep 拆分后的子串会先 trim space，若子串为空则会跳过
func ToSliceFunc[T any](str string, sep string, fn func(sub string) (T, error)) ([]T, error) {
	arr := strings.Split(str, sep)
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		num, err := fn(v)
		if err != nil {
			return nil, err
		}
		result = append(result, num)
	}
	return result, nil
}

func ToInts(str string, sep string) ([]int, error) {
	return ToSliceFunc[int](str, sep, strconv.Atoi)
}

func ToInt8s(str string, sep string) ([]int8, error) {
	return ToSliceFunc[int8](str, sep, func(sub string) (int8, error) {
		num, err := strconv.ParseInt(sub, 10, 8)
		return int8(num), err
	})
}

func ToInt16s(str string, sep string) ([]int16, error) {
	return ToSliceFunc[int16](str, sep, func(sub string) (int16, error) {
		num, err := strconv.ParseInt(sub, 10, 16)
		return int16(num), err
	})
}

func ToInt32s(str string, sep string) ([]int32, error) {
	return ToSliceFunc[int32](str, sep, func(sub string) (int32, error) {
		num, err := strconv.ParseInt(sub, 10, 32)
		return int32(num), err
	})
}

func ToInt64s(str string, sep string) ([]int64, error) {
	return ToSliceFunc[int64](str, sep, func(sub string) (int64, error) {
		return strconv.ParseInt(sub, 10, 64)
	})
}

func ToUints(str string, sep string) ([]uint, error) {
	return ToSliceFunc[uint](str, sep, func(sub string) (uint, error) {
		num, err := strconv.ParseUint(sub, 10, 0)
		return uint(num), err
	})
}

func ToUint8s(str string, sep string) ([]uint8, error) {
	return ToSliceFunc[uint8](str, sep, func(sub string) (uint8, error) {
		num, err := strconv.ParseUint(sub, 10, 8)
		return uint8(num), err
	})
}

func ToUint16s(str string, sep string) ([]uint16, error) {
	return ToSliceFunc[uint16](str, sep, func(sub string) (uint16, error) {
		num, err := strconv.ParseUint(sub, 10, 16)
		return uint16(num), err
	})
}

func ToUint32s(str string, sep string) ([]uint32, error) {
	return ToSliceFunc[uint32](str, sep, func(sub string) (uint32, error) {
		num, err := strconv.ParseUint(sub, 10, 32)
		return uint32(num), err
	})
}

func ToUint64s(str string, sep string) ([]uint64, error) {
	return ToSliceFunc[uint64](str, sep, func(sub string) (uint64, error) {
		return strconv.ParseUint(sub, 10, 64)
	})
}

// ToBools 将字符串 str 解析为 []bool
//
//	使用 sep 拆分后的子串会 trim space，若子串为空则会跳过。
//	子字符串  "1", "t", "T", "true", "TRUE", "IsTrue" 将解析为 true。
//	子字符串 "0", "f", "F", "false", "FALSE", "IsFalse" 将解析为 false。
//	其他子串会导致解析失败
func ToBools(str string, sep string) ([]bool, error) {
	return ToSliceFunc[bool](str, sep, strconv.ParseBool)
}

func ToFloat32s(str string, sep string) ([]float32, error) {
	return ToSliceFunc[float32](str, sep, func(sub string) (float32, error) {
		num, err := strconv.ParseFloat(sub, 32)
		return float32(num), err
	})
}

func ToFloat64s(str string, sep string) ([]float64, error) {
	return ToSliceFunc[float64](str, sep, func(sub string) (float64, error) {
		return strconv.ParseFloat(sub, 64)
	})
}

// ToStrings 使用 sep 将字符串拆分为 []string, 会对子串 trim space,并剔除掉空的子串
func ToStrings(str string, sep string) []string {
	values, _ := ToSliceFunc[string](str, sep, func(sub string) (string, error) {
		return sub, nil
	})
	return values
}
