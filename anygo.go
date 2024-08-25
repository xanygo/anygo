//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package anygo

// Ternary 三元表达式
func Ternary[T any](cond bool, trueValue T, falseValue T) T {
	if cond {
		return trueValue
	}
	return falseValue
}
