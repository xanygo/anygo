//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

func Get[Map ~map[K]V, K comparable, V any](m Map, key K) (v V, found bool) {
	if m == nil {
		return v, false
	}
	v, found = m[key]
	return v, found
}

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

func HasKey[Map ~map[K]V, K comparable, V any](m Map, key K) bool {
	if m == nil {
		return false
	}
	_, found := m[key]
	return found
}

func HasKeyValue[Map ~map[K]V, K comparable, V comparable](m Map, key K, value V) bool {
	if m == nil {
		return false
	}
	val, found := m[key]
	return found && val == value
}

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
