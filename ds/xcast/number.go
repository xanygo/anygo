//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-23

package xcast

import (
	"math"
	"strconv"
)

type (
	IntegerTypes interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}

	FloatTypes interface {
		~float32 | ~float64
	}
)

func ToInteger[T IntegerTypes](v any) T {
	z, _ := Integer[T](v)
	return z
}

// Integer 将任意类型 v 转换为目标整数类型 T。
// 转换失败或溢出返回 (0, false)。
func Integer[T IntegerTypes](v any) (T, bool) {
	var zero T

	switch x := v.(type) {
	// -------------------
	// 有符号整数类型
	// -------------------
	case int:
		return fromInt64[T](int64(x))
	case int8:
		return fromInt64[T](int64(x))
	case int16:
		return fromInt64[T](int64(x))
	case int32:
		return fromInt64[T](int64(x))
	case int64:
		return fromInt64[T](x)

	// -------------------
	// 无符号整数类型
	// -------------------
	case uint:
		return fromUint64[T](uint64(x))
	case uint8:
		return fromUint64[T](uint64(x))
	case uint16:
		return fromUint64[T](uint64(x))
	case uint32:
		return fromUint64[T](uint64(x))
	case uint64:
		return fromUint64[T](x)

	case float32:
		n, ok := floatToInt64(float64(x))
		if ok {
			return fromInt64[T](n)
		}
		return zero, false
	case float64:
		n, ok := floatToInt64(x)
		if ok {
			return fromInt64[T](n)
		}
		return zero, false

	// -------------------
	// 字符串类型
	// -------------------
	case string:
		if i, err := strconv.ParseInt(x, 10, 64); err == nil {
			return fromInt64[T](i)
		}
		if u, err := strconv.ParseUint(x, 10, 64); err == nil {
			return fromUint64[T](u)
		}
		return zero, false

	// -------------------
	// bool 转换
	// -------------------
	case bool:
		if x {
			return 1, true
		}
		return 0, true

	default:
		return zero, false
	}
}

// fromInt64 将 int64 安全转换为目标类型 T
func fromInt64[T IntegerTypes](v int64) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case int:
		if v < math.MinInt || v > math.MaxInt {
			return zero, false
		}
		return T(v), true
	case int8:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return zero, false
		}
		return T(v), true
	case int16:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return zero, false
		}
		return T(v), true
	case int32:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return zero, false
		}
		return T(v), true
	case int64:
		return T(v), true
	case uint, uint8, uint16, uint32, uint64:
		if v < 0 {
			return zero, false
		}
		return fromUint64[T](uint64(v))
	default:
		return zero, false
	}
}

// fromUint64 将 uint64 安全转换为目标类型 T
func fromUint64[T IntegerTypes](v uint64) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case int:
		if v > uint64(math.MaxInt) {
			return zero, false
		}
		return T(v), true
	case int8:
		if v > math.MaxInt8 {
			return zero, false
		}
		return T(v), true
	case int16:
		if v > math.MaxInt16 {
			return zero, false
		}
		return T(v), true
	case int32:
		if v > math.MaxInt32 {
			return zero, false
		}
		return T(v), true
	case int64:
		if v > math.MaxInt64 {
			return zero, false
		}
		return T(v), true
	case uint:
		if v > math.MaxUint {
			return zero, false
		}
		return T(v), true
	case uint8:
		if v > math.MaxUint8 {
			return zero, false
		}
		return T(v), true
	case uint16:
		if v > math.MaxUint16 {
			return zero, false
		}
		return T(v), true
	case uint32:
		if v > math.MaxUint32 {
			return zero, false
		}
		return T(v), true
	case uint64:
		return T(v), true
	default:
		return zero, false
	}
}

func floatToInt64(f float64) (int64, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	if f > math.MaxInt64 || f < math.MinInt64 {
		return 0, false
	}
	return int64(f), true
}

func ToFloat[T FloatTypes](v any) T {
	z, _ := Float[T](v)
	return z
}

// Float 将任意类型 v 转换为浮点类型 T（float32/float64）
// 转换失败返回 (0, false)
func Float[T FloatTypes](v any) (T, bool) {
	switch x := v.(type) {
	case float32:
		return T(x), true
	case float64:
		return T(x), true

	// 各类整数
	case int:
		return T(x), true
	case int8:
		return T(x), true
	case int16:
		return T(x), true
	case int32:
		return T(x), true
	case int64:
		return T(x), true
	case uint:
		return T(x), true
	case uint8:
		return T(x), true
	case uint16:
		return T(x), true
	case uint32:
		return T(x), true
	case uint64:
		return T(x), true

	case bool:
		if x {
			return T(1), true
		}
		return T(0), true

	case string:
		var t T
		switch any(t).(type) {
		case float32:
			f, err := strconv.ParseFloat(x, 32)
			return T(f), err == nil
		case float64:
			f, err := strconv.ParseFloat(x, 64)
			return T(f), err == nil
		}
	}
	return 0, false
}
