package zreflect_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestSliceContains(t *testing.T) {
	xt.True(t, zreflect.SliceContains([]int{1, 2, 3}, 1))

	var a any = []int{1, 2, 3}
	xt.True(t, zreflect.SliceContains(a, 1))

	xt.True(t, zreflect.SliceContains([]int{1, 2, 3}, 3))
	xt.True(t, zreflect.SliceContains([]int{1, 2, 3}, uint(3)))

	xt.True(t, zreflect.SliceContains([]int64{1, 2, 3}, 3))
	xt.True(t, zreflect.SliceContains([]uint{1, 2, 3}, 3))
	xt.True(t, zreflect.SliceContains([]uint{1, 2, 3}, uint(3)))

	xt.False(t, zreflect.SliceContains([]uint{1, 2, 3}, 4))
	xt.False(t, zreflect.SliceContains([]uint{1, 2, 3}, "4"))
	xt.False(t, zreflect.SliceContains([]uint{1, 2, 3}, "1"))
}

func TestRangeSlice(t *testing.T) {
	xt.NoError(t, zreflect.RangeSlice(nil, func(item string) error { return nil }))
	xt.Error(t, zreflect.RangeSlice("hello", func(item string) error { return nil }))
	var a []string
	xt.NoError(t, zreflect.RangeSlice(a, func(item string) error { return nil }))
	var b []any
	xt.NoError(t, zreflect.RangeSlice(b, func(item string) error { return nil }))

	c := []string{"a", "b"}
	var got1 []string
	xt.NoError(t, zreflect.RangeSlice(c, func(item string) error {
		got1 = append(got1, item)
		return nil
	}))
	xt.Equal(t, got1, c)

	d := []any{"a", "b"}
	var got2 []string
	xt.NoError(t, zreflect.RangeSlice(d, func(item string) error {
		got2 = append(got2, item)
		return nil
	}))
	xt.Equal(t, got2, []string{"a", "b"})

	e := any([]any{"a", "b"})
	var got3 []string
	xt.NoError(t, zreflect.RangeSlice(e, func(item string) error {
		got3 = append(got3, item)
		return nil
	}))
	xt.Equal(t, got3, []string{"a", "b"})
}
