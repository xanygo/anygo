//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package xslice

import "slices"

// DiffMore 查找到 new 相比 old 增量的部分，总是返回一个全新的 slice
func DiffMore[S ~[]E, E comparable](old, new S) S {
	if len(old) == 0 {
		if len(new) == 0 {
			return nil
		}
		return slices.Clone(new)
	}
	var result S
	om := ToMap(old, true)
	for _, v := range new {
		if !om[v] {
			result = append(result, v)
		}
	}
	return result
}
