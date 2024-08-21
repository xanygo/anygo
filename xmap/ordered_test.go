//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xmap_test

import (
	"fmt"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xmap"
)

func TestSorted(t *testing.T) {
	var m1 xmap.Ordered[int, string]
	fst.False(t, m1.Has(0))
	fst.False(t, m1.HasAny(0, 1, 2))
	m1.Set(0, "v0")
	m1.Set(1, "v1")
	fst.True(t, m1.Has(0))
	fst.True(t, m1.HasAny(3, 0))
	fst.Equal(t, map[int]string{0: "v0", 1: "v1"}, m1.Map(false))
	fst.Equal(t, map[int]string{0: "v0", 1: "v1"}, m1.Map(true))

	keys := []int{0, 1}
	for i := 2; i < 10; i++ {
		keys = append(keys, i)
		val := fmt.Sprintf("v_%d", i)
		m1.Set(i, val)
		fst.Equal(t, val, m1.MustGet(i))
	}
	fst.Equal(t, keys, m1.Keys())
	fst.Equal(t, 10, m1.Len())
	var keys1 []int
	m1.Range(func(key int, value string) bool {
		keys1 = append(keys1, key)
		return true
	})
	fst.Equal(t, keys, keys1)

	m1.Delete(1, 3)
	keys2 := []int{0, 2, 4, 5, 6, 7, 8, 9}
	fst.Equal(t, keys2, m1.Keys())

	fst.Equal(t, "v_2", m1.GetDefault(2, ""))
	fst.Equal(t, "v0", m1.GetDefault(0, ""))
	m1.Clear()
	fst.Equal(t, 0, m1.Len())
	fst.Empty(t, m1.Keys())
	fst.Empty(t, m1.Map(false))
	fst.Empty(t, m1.Map(true))
}
