//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/xanygo/anygo/internal/zslice"
)

// Merge merge 多个 slice 为一个，并最终返回一个新的 slice
func Merge[S ~[]T, T any](items ...S) S {
	return zslice.Merge(items...)
}

// SafeMerge 安全的合并两个Slice,并且若其中一个为空，另一个不为空时，直接返回不为空的
// 若两个都不为空，则返回一个全新的
func SafeMerge[S ~[]T, T any](a S, b S) S {
	return zslice.SafeMerge(a, b)
}

// Unique 返回去重后的 slice
func Unique[S ~[]T, T comparable](arr S) S {
	return zslice.Unique(arr)
}

// ContainsAny 判断 all 中是否包含 values 的任意一个值
func ContainsAny[E comparable](all []E, values ...E) bool {
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
func ToMap[E comparable, V any](s []E, value V) map[E]V {
	result := make(map[E]V, len(s))
	for i := 0; i < len(s); i++ {
		result[s[i]] = value
	}
	return result
}

// ToMapFunc 使用回调函数，将 slice 转换为 map
func ToMapFunc[E any, K comparable, V any](s []E, fn func(index int, v E) (K, V)) map[K]V {
	result := make(map[K]V, len(s))
	for i := 0; i < len(s); i++ {
		k, v := fn(i, s[i])
		result[k] = v
	}
	return result
}

// ToAnys 转换为 []any 类型
func ToAnys[E any](s []E) []any {
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
	return zslice.DeleteValue(s, values...)
}

// DeleteFuncN 使用 del 方法限量删除 limit 个元素
//
// 若 Limit <1 则为不限制删除个数
//
//	注意：该方法和标准库的 slices.DeleteFunc 一样，会影响传入的 S(删除的会被置空)
func DeleteFuncN[S ~[]E, E any](s S, del func(E) bool, limit int) S {
	i := slices.IndexFunc(s, del)
	if i == -1 {
		return s
	}
	deleted := 1
	// Don't start copying elements until we find one to delete.
	for j := i + 1; j < len(s); j++ {
		v := s[j]
		if del(v) && (limit < 1 || deleted < limit) {
			deleted++
			continue
		}
		s[i] = v
		i++
	}
	clear(s[i:]) // zero/nil out the obsolete elements, for GC
	return s[:i]
}

// RevIndexFunc  反向查找，若每找到，会返回 -1
func RevIndexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	for j := len(s) - 1; j > 0; j-- {
		if f(s[j]) {
			return j
		}
	}
	return -1
}

// RevDeleteFuncN 从尾端，使用 del 方法限量删除 limit 个元素
//
// 若 Limit <1 则为不限制删除个数
//
//	注意：该方法和标准库的 slices.DeleteFunc 一样，会影响传入的 S(删除的会被置空)
func RevDeleteFuncN[S ~[]E, E any](s S, del func(E) bool, limit int) S {
	i := RevIndexFunc(s, del)
	if i == -1 {
		return s
	}
	slices.Reverse(s)
	nl := DeleteFuncN(s, del, limit)
	slices.Reverse(nl)
	return nl
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
func JoinFunc[E any](arr []E, format func(val E) string, sep string) string {
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
func Join[E any](arr []E, sep string) string {
	return JoinFunc(arr, func(val E) string {
		var obj any = val
		if sg, ok := obj.(fmt.Stringer); ok {
			return sg.String()
		}
		return fmt.Sprint(val)
	}, sep)
}

// Filter 过滤删选出满足条件的元素
//
// filter: 过滤函数，参数依次为 index-元素索引、item 元素、okTotal-已过滤满足条件的个数
func Filter[S ~[]E, E any](arr S, filter func(index int, item E, okTotal int) bool) S {
	if len(arr) == 0 {
		return nil
	}
	var result S
	for i := 0; i < len(arr); i++ {
		if filter(i, arr[i], len(result)) {
			result = append(result, arr[i])
		}
	}
	return result
}

func FilterAs[E any, Y any](arr []E, filter func(index int, item E, ok int) (Y, bool)) []Y {
	if len(arr) == 0 {
		return nil
	}
	var result []Y
	for i := 0; i < len(arr); i++ {
		if item, ok := filter(i, arr[i], len(result)); ok {
			result = append(result, item)
		}
	}
	return result
}

func FilterOne[S ~[]E, E any](arr S, filter func(index int, item E) bool) (e E, ok bool) {
	if len(arr) == 0 {
		return e, false
	}
	for i := 0; i < len(arr); i++ {
		if filter(i, arr[i]) {
			return arr[i], true
		}
	}
	return e, false
}

// MapFilter 使用回调函数 fn 依次对 slice 的元素处理，fn 返回的第二个值为 false 则丢弃该值，否则更新
func MapFilter[S ~[]E, E any](arr S, fn func(index int, item E) (E, bool)) S {
	if len(arr) == 0 {
		return nil
	}
	result := make(S, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		if nv, ok := fn(i, arr[i]); ok {
			result = append(result, nv)
		}
	}
	return result
}

func MapFunc[S ~[]E, E any](arr S, fn func(item E) E) S {
	if len(arr) == 0 {
		return nil
	}
	result := make(S, len(arr))
	for i := 0; i < len(arr); i++ {
		result[i] = fn(arr[i])
	}
	return result
}

// CountFunc 统计 slice 中满足条件的元素个数
func CountFunc[S ~[]E, E any](arr S, fn func(index int, item E) bool) int64 {
	if len(arr) == 0 {
		return 0
	}
	var result int64
	for i := 0; i < len(arr); i++ {
		if fn(i, arr[i]) {
			result++
		}
	}
	return result
}

// Range 遍历任意类型的 slice 、Array，返回 value 满足条件且被 fn 接收的个数
func Range[T any](obj any, fn func(item T) bool) int {
	if obj == nil {
		return 0
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return 0
	}
	var cnt int
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i).Interface()
		val, ok := elem.(T)
		if !ok {
			continue
		}
		cnt++
		if !fn(val) {
			break
		}
	}
	return cnt
}

// Chunk 将 slice 拆分为同等大小 size 的多个子 slice
func Chunk[S ~[]E, E any](s S, size int) []S {
	if size <= 0 {
		return nil
	}
	chunks := make([]S, 0, len(s)/size+1)
	for i := 0; i < len(s); i += size {
		end := min(size, len(s[i:]))
		chunks = append(chunks, s[i:i+end:i+end])
	}
	return chunks
}

func Repeat[E any](v E, count int) []E {
	result := make([]E, count)
	for i := 0; i < count; i++ {
		result[i] = v
	}
	return result
}
