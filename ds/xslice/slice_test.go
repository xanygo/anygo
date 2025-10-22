//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

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

func TestContainsAny(t *testing.T) {
	xt.True(t, ContainsAny([]int{1, 3, 5}, 3))
	xt.True(t, ContainsAny([]int{1, 3, 5}, 6, 5))
	xt.False(t, ContainsAny([]int{1, 3, 5}, 6, 7))
	xt.False(t, ContainsAny([]int{1, 3, 5}))
	var a []int
	xt.False(t, ContainsAny(a, 1))
	xt.False(t, ContainsAny(a))
}

func TestToAnys(t *testing.T) {
	xt.Equal(t, []any{1}, ToAnys([]int{1}))
	var a []int
	xt.Nil(t, ToAnys(a))
}

func TestDeleteValue(t *testing.T) {
	xt.Equal(t, []int{1}, DeleteValue([]int{1}, 2))
	xt.Equal(t, []int{1}, DeleteValue([]int{1, 2}, 2))
	xt.Equal(t, []int{1}, DeleteValue([]int{1, 2, 2}, 2))
}

func TestPopHead(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1, ok1 := PopHead(a1)
		xt.False(t, ok1)
		xt.Equal(t, 0, v1)
		xt.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1, ok1 := PopHead(a1)
		xt.True(t, ok1)
		xt.Equal(t, 1, v1)
		xt.Empty(t, g1)

		g2, v2, ok2 := PopHead(g1)
		xt.False(t, ok2)
		xt.Equal(t, 0, v2)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1, ok1 := PopHead(a1)
		xt.True(t, ok1)
		xt.Equal(t, 1, v1)
		xt.NotEmpty(t, g1)

		g2, v2, ok2 := PopHead(g1)
		xt.True(t, ok2)
		xt.Equal(t, 2, v2)
		xt.Empty(t, g2)
	})
}

func TestPopHeadN(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1 := PopHeadN(a1, 2)
		xt.Empty(t, v1)
		xt.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1 := PopHeadN(a1, 2)
		xt.Equal(t, []int{1}, v1)
		xt.Empty(t, g1)

		g2, v2 := PopHeadN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1 := PopHeadN(a1, 2)
		xt.Equal(t, []int{1, 2}, v1)
		xt.Empty(t, g1)

		g2, v2 := PopHeadN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})
}

func TestPopTail(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1, ok1 := PopTail(a1)
		xt.False(t, ok1)
		xt.Equal(t, 0, v1)
		xt.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1, ok1 := PopTail(a1)
		xt.True(t, ok1)
		xt.Equal(t, 1, v1)
		xt.Empty(t, g1)

		g2, v2, ok2 := PopTail(g1)
		xt.False(t, ok2)
		xt.Equal(t, 0, v2)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1, ok1 := PopTail(a1)
		xt.True(t, ok1)
		xt.Equal(t, 2, v1)
		xt.NotEmpty(t, g1)

		g2, v2, ok2 := PopTail(g1)
		xt.True(t, ok2)
		xt.Equal(t, 1, v2)
		xt.Empty(t, g2)
	})
}

func TestPopTailN(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1 := PopTailN(a1, 2)
		xt.Empty(t, v1)
		xt.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1 := PopTailN(a1, 2)
		xt.Equal(t, []int{1}, v1)
		xt.Empty(t, g1)

		g2, v2 := PopTailN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1 := PopTailN(a1, 2)
		xt.Equal(t, []int{2, 1}, v1)
		xt.Empty(t, g1)

		g2, v2 := PopTailN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})
}

func TestJoin(t *testing.T) {
	xt.Equal(t, "1-2", Join([]int{1, 2}, "-"))
	xt.Equal(t, "", Join([]int{}, "-"))
}

func TestDeleteFuncN(t *testing.T) {
	arr := []int{1, 2, 3, 2, 2, 2, 2}
	got := DeleteFuncN(arr, func(i int) bool {
		return i == 2
	}, 3)
	want := []int{1, 3, 2, 2}
	xt.Equal(t, want, got)
}

func TestRevDeleteFuncN(t *testing.T) {
	arr1 := []int{1, 2, 3, 2, 2, 2, 2}
	got1 := RevDeleteFuncN(arr1, func(i int) bool {
		return i == 2
	}, 3)
	want1 := []int{1, 2, 3, 2}
	xt.Equal(t, want1, got1)

	arr1 = []int{1, 2, 3, 2, 2, 2, 2}
	got2 := RevDeleteFuncN(arr1, func(i int) bool {
		return i == 5
	}, 3)
	xt.Equal(t, arr1, got2)

	arr1 = []int{1, 2, 3, 2, 2, 2, 2}
	got3 := RevDeleteFuncN(arr1, func(i int) bool {
		return i == 2
	}, 0)
	want3 := []int{1, 3}
	xt.Equal(t, want3, got3)
}

func TestRange(t *testing.T) {
	s1 := []any{"1", 2, 3, int8(3)}
	var list1 []int
	num := Range[int](s1, func(item int) bool {
		list1 = append(list1, item)
		return true
	})
	xt.Equal(t, 2, num)
	xt.Equal(t, []int{2, 3}, list1)

	var list2 []int64
	num2 := Range[int64](s1, func(item int64) bool {
		list2 = append(list2, item)
		return true
	})
	xt.Empty(t, list2)
	xt.Equal(t, 0, num2)
}
