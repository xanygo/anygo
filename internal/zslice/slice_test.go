//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package zslice

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestMerge(t *testing.T) {
	xt.Equal(t, Merge([]int{}), []int{})
	xt.Equal(t, Merge([]int{1}, []int{2}), []int{1, 2})
}

func TestUnique(t *testing.T) {
	xt.Equal(t, Unique([]int{1, 1, 1}), []int{1})
	xt.Equal(t, Unique([]int{1, 2, 1}), []int{1, 2})
}

func TestDeleteValue(t *testing.T) {
	xt.Equal(t, DeleteValue([]int{1}, 2), []int{1})
	xt.Equal(t, DeleteValue([]int{1, 2}, 2), []int{1})
	xt.Equal(t, DeleteValue([]int{1, 2, 2}, 2), []int{1})
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
	xt.Equal(t, string(b1), "54321")

	b2 := []byte("1-")
	Reverse(b2)
	xt.Equal(t, string(b2), "-1")
}
