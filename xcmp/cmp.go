//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

//go:build go1.26

package xcmp

import "time"

var _ Comparable[time.Time] = time.Time{}

// Comparable  表示可以与同类型对象比较顺序的类型。
//
//	Compare 方法返回值：
//	-1 ：当前对象 < 参数对象(other)，
//	0  ：相等，
//	1  ：当前对象 > 参数对象(other)
type Comparable[T Comparable[T]] interface {
	Compare(other T) int
}

// OrderCustomAsc 升序排序规则，适用于实现了 Comparable 接口的类型
func OrderCustomAsc[T any, V Comparable[V]](get func(T) V) OrderRule[T] {
	return func(a, b T) int {
		return get(a).Compare(get(b))
	}
}

// OrderCustomDesc 降序排序规则，适用于实现了 Comparable 接口的类型
func OrderCustomDesc[T any, V Comparable[V]](get func(T) V) OrderRule[T] {
	return func(a, b T) int {
		return get(b).Compare(get(a))
	}
}
