//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package xslice

import "slices"

// Difference 返回 a 中不属于 b 的元素，总是返回一个全新的 slice。
func Difference[S ~[]E, E comparable](a, b S) S {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return slices.Clone(a)
	}
	var result S
	om := ToMap(b, struct{}{})
	for _, v := range a {
		if _, has := om[v]; !has {
			result = append(result, v)
		}
	}
	return result
}
