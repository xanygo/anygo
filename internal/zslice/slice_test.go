//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package zslice

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestMerge(t *testing.T) {
	xt.Equal(t, []int{}, Merge([]int{}))
	xt.Equal(t, []int{1, 2}, Merge([]int{1}, []int{2}))
}

func TestUnique(t *testing.T) {
	xt.Equal(t, []int{1}, Unique([]int{1, 1, 1}))
	xt.Equal(t, []int{1, 2}, Unique([]int{1, 2, 1}))
}

func TestDeleteValue(t *testing.T) {
	xt.Equal(t, []int{1}, DeleteValue([]int{1}, 2))
	xt.Equal(t, []int{1}, DeleteValue([]int{1, 2}, 2))
	xt.Equal(t, []int{1}, DeleteValue([]int{1, 2, 2}, 2))
}

func TestContainsAny(t *testing.T) {
	xt.True(t, ContainsAny([]int{1, 3, 5}, 3))
	xt.True(t, ContainsAny([]int{1, 3, 5}, 6, 5))
	xt.False(t, ContainsAny([]int{1, 3, 5}, 6, 7))
	xt.False(t, ContainsAny([]int{1, 3, 5}))
	var a []int
	xt.False(t, ContainsAny(a, 1))
	xt.False(t, ContainsAny(a))
}

func TestReverse(t *testing.T) {
	b1 := []byte("12345")
	Reverse(b1)
	xt.Equal(t, "54321", string(b1))

	b2 := []byte("1-")
	Reverse(b2)
	xt.Equal(t, "-1", string(b2))
}
