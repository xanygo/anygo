//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-27

package zslice

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestMerge(t *testing.T) {
	fst.Equal(t, []int{}, Merge([]int{}))
	fst.Equal(t, []int{1, 2}, Merge([]int{1}, []int{2}))
}

func TestUnique(t *testing.T) {
	fst.Equal(t, []int{1}, Unique([]int{1, 1, 1}))
	fst.Equal(t, []int{1, 2}, Unique([]int{1, 2, 1}))
}

func TestDeleteValue(t *testing.T) {
	fst.Equal(t, []int{1}, DeleteValue([]int{1}, 2))
	fst.Equal(t, []int{1}, DeleteValue([]int{1, 2}, 2))
	fst.Equal(t, []int{1}, DeleteValue([]int{1, 2, 2}, 2))
}

func TestContainsAny(t *testing.T) {
	fst.True(t, ContainsAny([]int{1, 3, 5}, 3))
	fst.True(t, ContainsAny([]int{1, 3, 5}, 6, 5))
	fst.False(t, ContainsAny([]int{1, 3, 5}, 6, 7))
	fst.False(t, ContainsAny([]int{1, 3, 5}))
	var a []int
	fst.False(t, ContainsAny(a, 1))
	fst.False(t, ContainsAny(a))
}
