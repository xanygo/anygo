//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

import (
	"fmt"
	"slices"
	"strings"
)

// Merge merge 多个 slice 为一个，并最终返回一个新的 slice
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

// ToAnys 转换为 []any 类型
func ToAnys[S ~[]E, E any](s S) []any {
	if len(s) == 0 {
		return nil
	}
	result := make([]any, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i]
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

// PopHead 弹出头部的一个元素，若 slice 为空会返回 false
func PopHead[S ~[]E, E any](s S) (new S, value E, ok bool) {
	if len(s) == 0 {
		return s, value, false
	}
	value = s[0]
	new = slices.Delete(s, 0, 1)
	return new, value, true
}

// PopHeadN 弹出头部的 n 个元素
func PopHeadN[S ~[]E, E any](s S, n int) (new S, values S) {
	if len(s) == 0 {
		return s, nil
	}
	if n >= len(s) {
		return nil, s
	}
	values = s[0:n]
	new = slices.Delete(s, 0, n)
	return new, values
}

// PopTail 弹出尾部的一个元素，若 slice 为空会返回 false
func PopTail[S ~[]E, E any](s S) (new S, value E, has bool) {
	if len(s) == 0 {
		var emp E
		return s, emp, false
	}
	index := len(s) - 1
	value = s[index]
	new = slices.Delete(s, index, index+1)
	return new, value, true
}

// PopTailN 弹出尾部的 n 个元素
func PopTailN[S ~[]E, E any](s S, n int) (new S, values S) {
	if len(s) == 0 {
		return s, nil
	}
	if n >= len(s) {
		slices.Reverse(s)
		return nil, s
	}
	index := len(s) - n
	values = s[index:]
	slices.Reverse(values)
	new = slices.Delete(s, index, len(s))
	return new, values
}

// JoinFunc 将 slice 使用特定的 format 方法转换为 string ，然后使用 sep 链接为字符串
func JoinFunc[S ~[]E, E any](arr S, format func(val E) string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	elems := make([]string, len(arr))
	for i := 0; i < len(arr); i++ {
		elems[i] = format(arr[i])
	}
	return strings.Join(elems, sep)
}

// Join 将 slice 使用默认的 format 方法转换为 string ，然后使用 sep 链接为字符串
//
// 若元素有实现 String()string 方法，会优先采用该值，否则会使用 fmt.Sprint
func Join[S ~[]E, E any](arr S, sep string) string {
	return JoinFunc(arr, func(val E) string {
		var obj any = val
		if sg, ok := obj.(fmt.Stringer); ok {
			return sg.String()
		}
		return fmt.Sprint(val)
	}, sep)
}
