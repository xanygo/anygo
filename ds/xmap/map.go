//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/xanygo/anygo/internal/zreflect"
)

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

func GetString[K comparable, V any](m map[K]V, key K) (string, bool) {
	if len(m) == 0 {
		return "", false
	}
	v, found := m[key]
	if !found {
		return "", false
	}
	return zreflect.BaseTypeToString(v)
}

func GetInt64[K comparable, V any](m map[K]V, key K) (int64, bool) {
	if len(m) == 0 {
		return 0, false
	}
	v, found := m[key]
	if !found {
		return 0, false
	}
	return zreflect.BaseTypeToInt64(v)
}

func GetInt[K comparable, V any](m map[K]V, key K) (int, bool) {
	num, err := GetInt64(m, key)
	return int(num), err
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

func KeyValues[K comparable, V any](m map[K]V) ([]K, []V) {
	if len(m) == 0 {
		return nil, nil
	}
	keys := make([]K, 0, len(m))
	values := make([]V, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
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

// MustCreate 使用 pairs 对，创建一个 map,若失败则 panic
func MustCreate[K comparable, V any](pairs ...any) map[K]V {
	result, err := Create[K, V](pairs)
	if err != nil {
		panic(err)
	}
	return result
}

// Create 传入 k-v 对，创建 map，若 传入 value == nil，则会当做零值
func Create[K comparable, V any](pairs ...any) (map[K]V, error) {
	total := len(pairs)
	if total%2 != 0 {
		return nil, fmt.Errorf("invalid map pairs, has an odd number (%d) of elements", total)
	}
	result := make(map[K]V, total/2)
	for i := 0; i < total; i += 2 {
		key := pairs[i]
		val := pairs[i+1]

		kt, ok1 := key.(K)
		if !ok1 {
			return nil, fmt.Errorf("key(%d)=%#v is not %T", i, key, kt)
		}
		var vt V
		if val != nil {
			var ok2 bool
			vt, ok2 = val.(V)
			if !ok2 {
				return nil, fmt.Errorf("map[%v]=%#v is not %T", key, val, vt)
			}
		}
		result[kt] = vt
	}
	return result, nil
}

func Join[K string, V any](m map[K]V, sep string) string {
	if len(m) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m))
	keys := Keys[K, V](m)
	slices.Sort(keys)
	for _, k := range keys {
		line := fmt.Sprintf("%s=%v", k, m[k])
		lines = append(lines, line)
	}
	return strings.Join(lines, sep)
}

// KeysMiss 找出 map 中缺少的 keys
func KeysMiss[K comparable, V any](mp map[K]V, keys []K) []K {
	var result []K
	for _, k := range keys {
		if _, ok := mp[k]; !ok {
			result = append(result, k)
		}
	}
	return result
}

// Range 遍历任意类型的 map，返回 key、value 满足条件而且被 fn 接收的个数
func Range[K comparable, V any](m any, fn func(key K, val V) bool) int {
	if m == nil {
		return 0
	}
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return 0
	}
	var cnt int
	for _, key := range v.MapKeys() {
		k, ok := key.Interface().(K)
		if !ok {
			continue
		}
		val, ok := v.MapIndex(key).Interface().(V)
		if !ok {
			continue
		}
		cnt++
		if !fn(k, val) {
			break
		}
	}
	return cnt
}
