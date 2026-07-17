package zreflect_test

import (
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
	"testing"
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
