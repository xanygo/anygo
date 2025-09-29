//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-01-04

package tplfn

import (
	"fmt"
	"reflect"
)

// MathAdd 支持混合类型的数字相加，若包含 float 类型，则结果返回 float64, 否则返回 int64 类型
func MathAdd(items ...any) (any, error) {
	if len(items) == 0 {
		return int64(0), nil
	}
	var floatResult float64
	var hasFloat bool
	var intResult int64
	for idx, v := range items {
		rv := reflect.ValueOf(v)
		if rv.CanFloat() {
			hasFloat = true
			floatResult += rv.Float()
		} else if rv.CanInt() {
			intResult += rv.Int()
		} else if rv.CanUint() {
			intResult += int64(rv.Uint())
		} else {
			return nil, fmt.Errorf("[%d]=%#v cannot add", idx, v)
		}
	}
	if hasFloat {
		return floatResult + float64(intResult), nil
	}
	return intResult, nil
}

// MathSub 减法
func MathSub(first any, items ...any) (any, error) {
	var floatResult float64
	var hasFloat bool
	var intResult int64

	rf := reflect.ValueOf(first)
	if rf.CanFloat() {
		hasFloat = true
		floatResult = rf.Float()
	} else if rf.CanInt() {
		intResult = rf.Int()
	} else if rf.CanUint() {
		intResult = int64(rf.Uint())
	}

	for idx, v := range items {
		rv := reflect.ValueOf(v)
		if rv.CanFloat() {
			hasFloat = true
			floatResult -= rv.Float()
		} else if rv.CanInt() {
			intResult -= rv.Int()
		} else if rv.CanUint() {
			intResult -= int64(rv.Uint())
		} else {
			return nil, fmt.Errorf("[%d]=%#v cannot sub", idx, v)
		}
	}
	if hasFloat {
		return floatResult + float64(intResult), nil
	}
	return intResult, nil
}

// MathMul 乘法
func MathMul(items ...any) (any, error) {
	if len(items) == 0 {
		return int64(0), nil
	}
	var floatResult float64 = 1
	var hasFloat bool
	var intResult int64 = 1

	for idx, v := range items {
		rv := reflect.ValueOf(v)
		if rv.CanFloat() {
			hasFloat = true
			floatResult *= rv.Float()
		} else if rv.CanInt() {
			intResult *= rv.Int()
		} else if rv.CanUint() {
			intResult *= int64(rv.Uint())
		} else {
			return nil, fmt.Errorf("[%d]=%#v cannot mul", idx, v)
		}
	}
	if hasFloat {
		return floatResult * float64(intResult), nil
	}
	return intResult, nil
}

func MathDiv(first any, items ...any) (float64, error) {
	var result float64

	rf := reflect.ValueOf(first)
	if rf.CanFloat() {
		result = rf.Float()
	} else if rf.CanInt() {
		result = float64(rf.Int())
	} else if rf.CanUint() {
		result = float64(rf.Uint())
	}

	for idx, item := range items {
		if result == 0 {
			return float64(0), nil
		}
		rv := reflect.ValueOf(item)
		var dv float64
		if rv.CanFloat() {
			dv = rv.Float()
		} else if rv.CanInt() {
			dv = float64(rv.Int())
		} else if rv.CanUint() {
			dv = float64(rv.Uint())
		} else {
			return 0, fmt.Errorf("[%d]=%#v not number, cannot be divided", idx, item)
		}
		if dv == 0 {
			return 0, fmt.Errorf("[%d]=%#v is 0, cannot be divided ", idx, item)
		}
		result /= dv
	}
	return result, nil
}

func MathPercent(val float64) string {
	if val == 0 {
		return "0%"
	} else if val == 1 {
		return "100%"
	}
	return fmt.Sprintf("%.3f%%", val*100)
}

func MathComplement(val float64) string {
	if val == 0 {
		return "100%"
	} else if val == 1 {
		return "0%"
	}
	return fmt.Sprintf("%.3f%%", (1-val)*100)
}
