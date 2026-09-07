//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-06

package xcmp

import (
	"cmp"
)

// OrderRule 表示一条排序规则，用于确定两个元素在最终有序序列中的相对位置,用于 slices.SortFunc
//
// OrderRule 接收两个同类型元素 a 和 b，并返回一个整数：
//   - 返回值 -1 ( < 0 )：表示 a 应排在 b 前面
//   - 返回值  1 ( > 0 )：表示 a 应排在 b 后面
//   - 返回值  0 ：表示 a 与 b 在该规则下等价
//
// 并可通过组合多个 OrderRule 来表达复杂的排序优先级。
type OrderRule[T any] func(a, b T) int

// Chain 可用于 slices.SortFunc ，多条件排序
func Chain[T any](cmps ...OrderRule[T]) OrderRule[T] {
	return func(a, b T) int {
		for _, cmp := range cmps {
			if r := cmp(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
}

// ReverseChain 可用于 slices.SortFunc ，多条件排序,逆序
func ReverseChain[T any](cmps ...OrderRule[T]) OrderRule[T] {
	return func(a, b T) int {
		for _, cmp := range cmps {
			if r := cmp(a, b); r != 0 {
				if r > 0 {
					return -1
				}
				return 1
			}
		}
		return 0
	}
}

// TrueFront 排序规则：返回值为 true 时，则排在前面，否则排到后面
func TrueFront[T any](get func(T) bool) OrderRule[T] {
	return func(a, b T) int {
		return boolCmp(get(a), get(b), true)
	}
}

// TrueBack 排序规则：返回值为 true 时，则排在后面，否则排到前面
func TrueBack[T any](get func(T) bool) OrderRule[T] {
	return func(a, b T) int {
		return boolCmp(get(a), get(b), false)
	}
}

// OrderAsc 按值升序，小的排在前面
func OrderAsc[T any, V cmp.Ordered](get func(T) V) OrderRule[T] {
	return func(a, b T) int {
		av, bv := get(a), get(b)
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		default:
			return 0
		}
	}
}

// OrderDesc 按值降序，大的排在前面
func OrderDesc[T any, V cmp.Ordered](get func(T) V) OrderRule[T] {
	return func(a, b T) int {
		av, bv := get(a), get(b)
		switch {
		case av > bv:
			return -1
		case av < bv:
			return 1
		default:
			return 0
		}
	}
}

func boolCmp(a, b bool, before bool) int {
	if a == b {
		return 0
	}
	if before {
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
