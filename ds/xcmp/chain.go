//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-06

package xcmp

import (
	"bytes"
	"cmp"
	"strings"
)

type Comparator[T any] func(a, b T) int

// Chain 可用于 slices.SortFunc ，多条件排序
func Chain[T any](cmps ...Comparator[T]) Comparator[T] {
	return func(a, b T) int {
		for _, cmp := range cmps {
			if r := cmp(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
}

// StrContains 排序规则：字符串包含时，若 toHead == true，则排在前面
func StrContains[T any](get func(T) string, keyword string, toHead bool) Comparator[T] {
	return func(a, b T) int {
		av := strings.Contains(get(a), keyword)
		bv := strings.Contains(get(b), keyword)
		return boolCmp(av, bv, toHead)
	}
}

// BytesContains 排序规则：字符串包含时，若 toHead == true，则排在前面
func BytesContains[T any](get func(T) []byte, keyword []byte, toHead bool) Comparator[T] {
	return func(a, b T) int {
		av := bytes.Contains(get(a), keyword)
		bv := bytes.Contains(get(b), keyword)
		return boolCmp(av, bv, toHead)
	}
}

// Equal 排序规则：值等于 value 时，若 toHead == true，则排在前面
func Equal[T any, V comparable](get func(T) V, value V, toHead bool) Comparator[T] {
	return func(a, b T) int {
		av := get(a) == value
		bv := get(b) == value
		return boolCmp(av, bv, toHead)
	}
}

// Compare 排序规则：按照值的大小比较，若 asc=true，则升序；若asc=false, 则降序
func Compare[T any, V cmp.Ordered](get func(T) V, asc bool) Comparator[T] {
	return func(a, b T) int {
		av, bv := get(a), get(b)
		if av == bv {
			return 0
		}
		if asc {
			if av < bv {
				return -1
			}
			return 1
		}
		if av > bv {
			return -1
		}
		return 1
	}
}

func boolCmp(a, b bool, asc bool) int {
	if a == b {
		return 0
	}
	if asc {
		if a {
			return -1
		}
		return 1
	}
	if a {
		return 1
	}
	return -1
}
