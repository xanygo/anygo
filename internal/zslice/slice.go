//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package zslice

import "slices"

func Merge[S ~[]T, T any](items ...S) S {
	switch len(items) {
	case 0:
		return nil
	case 1:
		return slices.Clone(items[0])
	}

	var n int
	for i := 0; i < len(items); i++ {
		n += len(items[i])
	}
	cp := make([]T, 0, n)
	for i := 0; i < len(items); i++ {
		cp = append(cp, items[i]...)
	}
	return cp
}

// SafeMerge 安全的合并两个Slice,并且若其中一个为空，另一个不为空时，直接返回不为空的
// 若两个都不为空，则返回一个全新的
func SafeMerge[S ~[]T, T any](a S, b S) S {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	c := make(S, 0, len(a)+len(b))
	c = append(c, a...)
	c = append(c, b...)
	return c
}

// Unique 返回去重后的 slice
func Unique[S ~[]T, T comparable](arr S) S {
	if len(arr) < 2 {
		return arr
	}
	c := make(map[T]struct{}, len(arr))
	result := make([]T, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		v := arr[i]
		if _, ok := c[v]; ok {
			continue
		}
		c[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// DeleteValue 删除指定的值
func DeleteValue[S ~[]E, E comparable](s S, values ...E) S {
	if len(s) == 0 || len(values) == 0 {
		return s
	}
	kv := make(map[E]struct{}, len(values))
	for _, v := range values {
		kv[v] = struct{}{}
	}
	oldLen := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := kv[s[i]]; ok {
			s = append(s[:i], s[i+1:]...)
		}
	}
	clear(s[len(s):oldLen])
	return s
}

// ContainsAny 判断 all 中是否包含 values 的任意一个值
func ContainsAny[S ~[]E, E comparable](all S, values ...E) bool {
	if len(all) == 0 || len(values) == 0 {
		return false
	}
	kv := ToMap(values, struct{}{})
	for _, v := range all {
		if _, ok := kv[v]; ok {
			return true
		}
	}
	return false
}

// ToMap 转换为 map，map 的 key 是 slice 的值， map 的 value 是传入的 value
// 总是返回不为 nil 的 map
func ToMap[S ~[]E, E comparable, V any](s S, value V) map[E]V {
	result := make(map[E]V, len(s))
	for i := 0; i < len(s); i++ {
		result[s[i]] = value
	}
	return result
}

// Reverse 顺序反转
func Reverse[S ~[]E, E any](b S) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
