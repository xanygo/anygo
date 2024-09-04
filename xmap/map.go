//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

// Get 从 map 中读取指定 key 的值。支持 map 为 nil。
func Get[Map ~map[K]V, K comparable, V any](m Map, key K) (v V, found bool) {
	if m == nil {
		return v, false
	}
	v, found = m[key]
	return v, found
}

// GetDf  从 map 中读取指定 key 的值,若 key 不存在则返回默认值。支持 map 为 nil。
func GetDf[Map ~map[K]V, K comparable, V any](m Map, key K, def V) V {
	if m == nil {
		return def
	}
	v, found := m[key]
	if found {
		return v
	}
	return def
}

// HasKey 判断 map 中是否存在特定 key。支持 map 为 nil。
func HasKey[Map ~map[K]V, K comparable, V any](m Map, key K) bool {
	if m == nil {
		return false
	}
	_, found := m[key]
	return found
}

// HasKeyValue 判断 map 中是否有指定的 key 和 value。支持 map 为 nil。
func HasKeyValue[Map ~map[K]V, K comparable, V comparable](m Map, key K, value V) bool {
	if m == nil {
		return false
	}
	val, found := m[key]
	return found && val == value
}

// Keys 返回 map 所有的 key。支持 map 为 nil。
func Keys[Map ~map[K]V, K comparable, V any](m Map) []K {
	if len(m) == 0 {
		return nil
	}
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回 map 所有的 value。支持 map 为 nil。
func Values[Map ~map[K]V, K comparable, V any](m Map) []V {
	if len(m) == 0 {
		return nil
	}
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Filter  从 map 中 过滤出满足条件的项
//
// filter: 过滤函数，参数依次为 k、v 分别是 map 的 key 和 value、ok-已过滤满足条件的个数
func Filter[Map ~map[K]V, K comparable, V any](m Map, filter func(k K, v V, ok int) bool) Map {
	result := make(Map)
	if len(m) == 0 {
		return result
	}
	for k, v := range m {
		if filter(k, v, len(result)) {
			result[k] = v
		}
	}
	return result
}

// FilterKeys 从 map 中 过滤出满足条件的 key 列表
//
// filter: 过滤函数，参数依次为 k、v 分别是 map 的 key 和 value、ok-已过滤满足条件的个数
func FilterKeys[Map ~map[K]V, K comparable, V any](m Map, filter func(k K, v V, ok int) bool) []K {
	if len(m) == 0 {
		return nil
	}
	values := make([]K, 0, len(m))
	for k, v := range m {
		if filter(k, v, len(values)) {
			values = append(values, k)
		}
	}
	return values
}

// FilterValues 从 map 中 过滤出满足条件的 value 列表
//
// filter: 过滤函数，参数依次为 k、v 分别是 map 的 key 和 value、ok-已过滤满足条件的个数
func FilterValues[Map ~map[K]V, K comparable, V any](m Map, filter func(k K, v V, ok int) bool) []V {
	if len(m) == 0 {
		return nil
	}
	values := make([]V, 0, len(m))
	for k, v := range m {
		if filter(k, v, len(values)) {
			values = append(values, v)
		}
	}
	return values
}
