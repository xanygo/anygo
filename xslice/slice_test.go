//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

import (
	"slices"
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
	xt.Equal(t, ToAnys([]int{1}), []any{1})
	var a []int
	xt.Nil(t, ToAnys(a))
}

func TestDeleteValue(t *testing.T) {
	xt.Equal(t, DeleteValue([]int{1}, 2), []int{1})
	xt.Equal(t, DeleteValue([]int{1, 2}, 2), []int{1})
	xt.Equal(t, DeleteValue([]int{1, 2, 2}, 2), []int{1})
}

func TestPopHead(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1, ok1 := PopHead(a1)
		xt.False(t, ok1)
		xt.Equal(t, v1, 0)
		xt.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1, ok1 := PopHead(a1)
		xt.True(t, ok1)
		xt.Equal(t, v1, 1)
		xt.Empty(t, g1)

		g2, v2, ok2 := PopHead(g1)
		xt.False(t, ok2)
		xt.Equal(t, v2, 0)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1, ok1 := PopHead(a1)
		xt.True(t, ok1)
		xt.Equal(t, v1, 1)
		xt.Equal(t, g1, []int{2})

		g2, v2, ok2 := PopHead(g1)
		xt.True(t, ok2)
		xt.Equal(t, v2, 2)
		xt.Empty(t, g2)

		// 检查元slice长度未变化
		xt.Len(t, a1, 2)
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
		xt.Equal(t, v1, []int{1})
		xt.Empty(t, g1)

		g2, v2 := PopHeadN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1 := PopHeadN(a1, 2)
		xt.Equal(t, v1, []int{1, 2})
		xt.Empty(t, g1)

		g2, v2 := PopHeadN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})
	t.Run("len 5", func(t *testing.T) {
		a1 := []int{1, 2, 3, 4, 5}
		g1, v1 := PopHeadN(a1, 2)
		xt.Equal(t, v1, []int{1, 2})
		xt.Equal(t, g1, []int{3, 4, 5})

		g2, v2 := PopHeadN(g1, 2)
		xt.Equal(t, v2, []int{3, 4})
		xt.Equal(t, g2, []int{5})

		g3, v3 := PopHeadN(g2, 2)
		xt.Equal(t, v3, []int{5})
		xt.Empty(t, g3)
	})
}

func TestPopTail(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1, ok1 := PopTail(a1)
		xt.False(t, ok1)
		xt.Equal(t, v1, 0)
		xt.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1, ok1 := PopTail(a1)
		xt.True(t, ok1)
		xt.Equal(t, v1, 1)
		xt.Empty(t, g1)

		g2, v2, ok2 := PopTail(g1)
		xt.False(t, ok2)
		xt.Equal(t, v2, 0)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1, ok1 := PopTail(a1)
		xt.True(t, ok1)
		xt.Equal(t, v1, 2)
		xt.NotEmpty(t, g1)

		g2, v2, ok2 := PopTail(g1)
		xt.True(t, ok2)
		xt.Equal(t, v2, 1)
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
		xt.Equal(t, v1, []int{1})
		xt.Empty(t, g1)

		g2, v2 := PopTailN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1 := PopTailN(a1, 2)
		xt.Equal(t, v1, []int{2, 1})
		xt.Empty(t, g1)

		g2, v2 := PopTailN(g1, 2)
		xt.Empty(t, v2)
		xt.Empty(t, g2)
	})
	t.Run("len 5", func(t *testing.T) {
		a1 := []int{1, 2, 3, 4, 5}
		g1, v1 := PopTailN(a1, 2)
		xt.Equal(t, v1, []int{5, 4})
		xt.Equal(t, g1, []int{1, 2, 3})

		g2, v2 := PopTailN(g1, 2)
		xt.Equal(t, v2, []int{3, 2})
		xt.Equal(t, g2, []int{1})

		g3, v3 := PopTailN(g2, 2)
		xt.Equal(t, v3, []int{1})
		xt.Empty(t, g3)
	})
}

func TestJoin(t *testing.T) {
	xt.Equal(t, Join([]int{1, 2}, "-"), "1-2")
	xt.Equal(t, Join([]int{}, "-"), "")
}

func TestDeleteFuncN(t *testing.T) {
	arr := []int{1, 2, 3, 2, 2, 2, 2}
	got := DeleteFuncN(arr, func(i int) bool {
		return i == 2
	}, 3)
	want := []int{1, 3, 2, 2}
	xt.Equal(t, got, want)
}

func TestRevDeleteFuncN(t *testing.T) {
	arr1 := []int{1, 2, 3, 2, 2, 2, 2}
	got1 := RevDeleteFuncN(arr1, func(i int) bool {
		return i == 2
	}, 3)
	want1 := []int{1, 2, 3, 2}
	xt.Equal(t, got1, want1)

	arr1 = []int{1, 2, 3, 2, 2, 2, 2}
	got2 := RevDeleteFuncN(arr1, func(i int) bool {
		return i == 5
	}, 3)
	xt.Equal(t, got2, arr1)

	arr1 = []int{1, 2, 3, 2, 2, 2, 2}
	got3 := RevDeleteFuncN(arr1, func(i int) bool {
		return i == 2
	}, 0)
	want3 := []int{1, 3}
	xt.Equal(t, got3, want3)
}

func TestRange(t *testing.T) {
	s1 := []any{"1", 2, 3, int8(3)}
	var list1 []int
	ok := Range[int](s1, func(item int) bool {
		list1 = append(list1, item)
		return true
	})
	xt.True(t, ok)
	xt.Equal(t, list1, []int{2, 3})

	var list2 []int64
	ok = Range[int64](s1, func(item int64) bool {
		list2 = append(list2, item)
		return true
	})
	xt.True(t, ok)
	xt.Empty(t, list2)
}

func TestChunk(t *testing.T) {
	l1 := []int{1, 2, 3, 4, 5}
	got := Chunk(l1, 2)
	want := [][]int{
		{1, 2},
		{3, 4},
		{5},
	}
	xt.Equal(t, got, want)
}

func TestAllContains(t *testing.T) {
	miss, ok := AllContains([]string{"a", "b"}, []string{"a"})
	xt.True(t, ok)
	xt.Empty(t, miss)

	miss, ok = AllContains([]string{"a", "b"}, []string{"c"})
	xt.False(t, ok)
	xt.Equal(t, miss, "c")
}

func TestPopRand(t *testing.T) {
	l1 := []int{1, 2, 3, 4, 5}
	t.Run("pop1", func(t *testing.T) {
		var poped []int
		ns := slices.Clone(l1)
		arr := ns
		for i := 0; i < 5; i++ {
			var one int
			var ok bool
			arr, one, ok = PopRand(arr)
			xt.True(t, ok)
			xt.Len(t, arr, 4-i)
			xt.SliceContains(t, l1, one)
			poped = append(poped, one)
		}
		xt.Empty(t, arr)
		xt.SliceSortEqual(t, l1, poped)
		xt.Len(t, ns, 5)
	})
}

func TestPopRandN(t *testing.T) {
	l1 := []int{1, 2, 3, 4, 5}
	t.Run("pop1", func(t *testing.T) {
		var poped []int
		ns := slices.Clone(l1)
		var vs []int
		for i := 0; i < 2; i++ {
			ns, vs = PopRandN(ns, 2)
			xt.Len(t, ns, 3-i*2)
			xt.Len(t, vs, 2)
			xt.SliceContains(t, l1, vs...)
			poped = append(poped, vs...)
		}

		ns, vs = PopRandN(ns, 2)
		xt.Len(t, ns, 0)
		xt.Len(t, vs, 1)
		xt.SliceContains(t, l1, vs...)
		poped = append(poped, vs...)

		xt.Empty(t, ns)
		xt.SliceSortEqual(t, l1, poped)
	})
}
