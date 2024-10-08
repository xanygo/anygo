//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

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

func TestContainsAny(t *testing.T) {
	fst.True(t, ContainsAny([]int{1, 3, 5}, 3))
	fst.True(t, ContainsAny([]int{1, 3, 5}, 6, 5))
	fst.False(t, ContainsAny([]int{1, 3, 5}, 6, 7))
	fst.False(t, ContainsAny([]int{1, 3, 5}))
	var a []int
	fst.False(t, ContainsAny(a, 1))
	fst.False(t, ContainsAny(a))
}

func TestToAnys(t *testing.T) {
	fst.Equal(t, []any{1}, ToAnys([]int{1}))
	var a []int
	fst.Nil(t, ToAnys(a))
}

func TestDeleteValue(t *testing.T) {
	fst.Equal(t, []int{1}, DeleteValue([]int{1}, 2))
	fst.Equal(t, []int{1}, DeleteValue([]int{1, 2}, 2))
	fst.Equal(t, []int{1}, DeleteValue([]int{1, 2, 2}, 2))
}

func TestPopHead(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1, ok1 := PopHead(a1)
		fst.False(t, ok1)
		fst.Equal(t, 0, v1)
		fst.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1, ok1 := PopHead(a1)
		fst.True(t, ok1)
		fst.Equal(t, 1, v1)
		fst.Empty(t, g1)

		g2, v2, ok2 := PopHead(g1)
		fst.False(t, ok2)
		fst.Equal(t, 0, v2)
		fst.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1, ok1 := PopHead(a1)
		fst.True(t, ok1)
		fst.Equal(t, 1, v1)
		fst.NotEmpty(t, g1)

		g2, v2, ok2 := PopHead(g1)
		fst.True(t, ok2)
		fst.Equal(t, 2, v2)
		fst.Empty(t, g2)
	})
}

func TestPopHeadN(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1 := PopHeadN(a1, 2)
		fst.Empty(t, v1)
		fst.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1 := PopHeadN(a1, 2)
		fst.Equal(t, []int{1}, v1)
		fst.Empty(t, g1)

		g2, v2 := PopHeadN(g1, 2)
		fst.Empty(t, v2)
		fst.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1 := PopHeadN(a1, 2)
		fst.Equal(t, []int{1, 2}, v1)
		fst.Empty(t, g1)

		g2, v2 := PopHeadN(g1, 2)
		fst.Empty(t, v2)
		fst.Empty(t, g2)
	})
}

func TestPopTail(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1, ok1 := PopTail(a1)
		fst.False(t, ok1)
		fst.Equal(t, 0, v1)
		fst.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1, ok1 := PopTail(a1)
		fst.True(t, ok1)
		fst.Equal(t, 1, v1)
		fst.Empty(t, g1)

		g2, v2, ok2 := PopTail(g1)
		fst.False(t, ok2)
		fst.Equal(t, 0, v2)
		fst.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1, ok1 := PopTail(a1)
		fst.True(t, ok1)
		fst.Equal(t, 2, v1)
		fst.NotEmpty(t, g1)

		g2, v2, ok2 := PopTail(g1)
		fst.True(t, ok2)
		fst.Equal(t, 1, v2)
		fst.Empty(t, g2)
	})
}

func TestPopTailN(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var a1 []int
		g1, v1 := PopTailN(a1, 2)
		fst.Empty(t, v1)
		fst.Empty(t, g1)
	})

	t.Run("len 1", func(t *testing.T) {
		a1 := []int{1}
		g1, v1 := PopTailN(a1, 2)
		fst.Equal(t, []int{1}, v1)
		fst.Empty(t, g1)

		g2, v2 := PopTailN(g1, 2)
		fst.Empty(t, v2)
		fst.Empty(t, g2)
	})

	t.Run("len 2", func(t *testing.T) {
		a1 := []int{1, 2}
		g1, v1 := PopTailN(a1, 2)
		fst.Equal(t, []int{2, 1}, v1)
		fst.Empty(t, g1)

		g2, v2 := PopTailN(g1, 2)
		fst.Empty(t, v2)
		fst.Empty(t, g2)
	})
}

func TestJoin(t *testing.T) {
	fst.Equal(t, "1-2", Join([]int{1, 2}, "-"))
	fst.Equal(t, "", Join([]int{}, "-"))
}
