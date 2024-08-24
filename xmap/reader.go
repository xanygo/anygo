//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

func Get[K comparable, V any](m map[K]V, key K) (v V, found bool) {
	if m == nil {
		return v, false
	}
	v, found = m[key]
	return v, found
}

func GetDf[K comparable, V any](m map[K]V, key K, def V) V {
	if m == nil {
		return def
	}
	v, found := m[key]
	if found {
		return v
	}
	return def
}
