//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package zslice

func Merge[S ~[]T, T any](items ...S) S {
	var n int
	for i := range items {
		n += len(items[i])
	}
	if n == 0 {
		return nil
	}
	cp := make([]T, 0, n)
	for i := range items {
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

// Get 从 s 中安全地读取索引 index 对应的值。
//
//	index >= 0 时从切片头部开始索引。
//	index < 0 时从切片尾部开始索引，-1 表示最后一个元素，-2 表示倒数第二个元素。
//	当 index 超出有效范围时返回零值和 false。
func Get[S ~[]E, E any](s S, index int) (E, bool) {
	var zero E
	if index < 0 {
		index += len(s)
	}

	if index < 0 || index >= len(s) {
		return zero, false
	}

	return s[index], true
}
