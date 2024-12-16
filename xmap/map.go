//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

// Get 从 map 中读取指定 key 的值。支持 map 为 nil。
func Get[K comparable, V any](m map[K]V, key K) (v V, found bool) {
	if len(m) == 0 {
		return v, false
	}
	v, found = m[key]
	return v, found
}

// GetDf  从 map 中读取指定 key 的值,若 key 不存在则返回默认值。支持 map 为 nil。
func GetDf[K comparable, V any](m map[K]V, key K, def V) V {
	if len(m) == 0 {
		return def
	}
	v, found := m[key]
	if found {
		return v
	}
	return def
}

// HasKey 判断 map 中是否存在特定 key。支持 map 为 nil。
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	if len(m) == 0 {
		return false
	}
	_, found := m[key]
	return found
}

// HasAnyKey 判断 map 中是否存在任意 key。支持 map 为 nil。
// 若 map 或者 keys 为空，均会返回 false
func HasAnyKey[K comparable, V any](m map[K]V, keys ...K) bool {
	if len(m) == 0 || len(keys) == 0 {
		return false
	}
	for _, key := range keys {
		if _, found := m[key]; found {
			return true
		}
	}

	return false
}

// HasKeyValue 判断 map 中是否有指定的 key 和 value。支持 map 为 nil。
func HasKeyValue[K comparable, V comparable](m map[K]V, key K, value V) bool {
	if len(m) == 0 {
		return false
	}
	val, found := m[key]
	return found && val == value
}

// HasAnyKeyValue 查找 map 中是否有 search 中的任意一项 key-value 。
// 若 m 或 search 为空，均会返回 false
func HasAnyKeyValue[K comparable, V comparable](m map[K]V, search map[K]V) bool {
	if len(m) == 0 || len(search) == 0 {
		return false
	}
	for key, value := range search {
		val, found := m[key]
		if found && val == value {
			return true
		}
	}
	return false
}

// Keys 返回 map 所有的 key。支持 map 为 nil。
func Keys[K comparable, V any](m map[K]V) []K {
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
func Values[K comparable, V any](m map[K]V) []V {
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

// FilterByKeys 从 map 中筛选出指定的 key 的项
func FilterByKeys[Map ~map[K]V, K comparable, V any](m Map, keys ...K) Map {
	result := make(Map)
	if len(m) == 0 || len(keys) == 0 {
		return result
	}
	km := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		km[k] = struct{}{}
	}
	for k, v := range m {
		if _, ok := km[k]; ok {
			result[k] = v
		}
	}
	return result
}

// FilterByValues 从 map 中筛选出指定的 key 的项
func FilterByValues[Map ~map[K]V, K comparable, V comparable](m Map, values ...V) Map {
	result := make(Map)
	if len(m) == 0 || len(values) == 0 {
		return result
	}
	vm := make(map[V]struct{}, len(values))
	for _, v := range values {
		vm[v] = struct{}{}
	}
	for k, v := range m {
		if _, ok := vm[v]; ok {
			result[k] = v
		}
	}
	return result
}

// FilterKeys 从 map 中 过滤出满足条件的 key 列表
//
// filter: 过滤函数，参数依次为 k、v 分别是 map 的 key 和 value、ok-已过滤满足条件的个数
func FilterKeys[K comparable, V any](m map[K]V, filter func(k K, v V, ok int) bool) []K {
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
func FilterValues[K comparable, V any](m map[K]V, filter func(k K, v V, ok int) bool) []V {
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

// Create 使用 pairs 对，创建一个 map,若 key 或者 value 为nil，则会跳过此条数据
func Create(pairs ...any) map[any]any {
	if len(pairs)%2 != 0 {
		panic("invalid map pairs")
	}
	result := make(map[any]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i]
		val := pairs[i+1]
		if key == nil || val == nil {
			continue
		}
		result[key] = val
	}
	return result
}
