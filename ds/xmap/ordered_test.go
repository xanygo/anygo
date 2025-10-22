//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-20

package xmap_test

import (
	"fmt"
	"iter"
	"testing"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xt"
)

type testOrderedType[K comparable, V any] interface {
	Set(key K, value V)
	Delete(keys ...K)
	Get(key K) (V, bool)
	GetDf(key K, def V) V
	MustGet(key K) V
	Has(key K) bool
	HasAny(keys ...K) bool
	Range(fn func(key K, value V) bool)
	Iter() iter.Seq2[K, V]
	Keys() []K
	Map(clone bool) map[K]V
	KVs(clone bool) ([]K, map[K]V)
	Values() []V
	Len() int
	Clear()
}

var (
	_ testOrderedType[int, int] = (*xmap.Ordered[int, int])(nil)
	_ testOrderedType[int, int] = (*xmap.OrderedSync[int, int])(nil)
)

func TestOrdered(t *testing.T) {
	check := func(t *testing.T, m1 testOrderedType[int, string]) {
		xt.False(t, m1.Has(0))
		xt.False(t, m1.HasAny(0, 1, 2))
		m1.Set(0, "v0")
		m1.Set(1, "v1")
		xt.True(t, m1.Has(0))
		xt.True(t, m1.HasAny(3, 0))
		xt.Equal(t, map[int]string{0: "v0", 1: "v1"}, m1.Map(false))
		xt.Equal(t, map[int]string{0: "v0", 1: "v1"}, m1.Map(true))

		keys := []int{0, 1}
		for i := 2; i < 10; i++ {
			keys = append(keys, i)
			val := fmt.Sprintf("v_%d", i)
			m1.Set(i, val)
			xt.Equal(t, val, m1.MustGet(i))
		}
		xt.Equal(t, keys, m1.Keys())
		xt.Equal(t, 10, m1.Len())
		var keys1 []int
		m1.Range(func(key int, value string) bool {
			keys1 = append(keys1, key)
			return true
		})
		xt.Equal(t, keys, keys1)

		m1.Delete(1, 3)
		keys2 := []int{0, 2, 4, 5, 6, 7, 8, 9}
		xt.Equal(t, keys2, m1.Keys())

		xt.Equal(t, "v_2", m1.GetDf(2, ""))
		xt.Equal(t, "v0", m1.GetDf(0, ""))
		m1.Clear()
		xt.Equal(t, 0, m1.Len())
		xt.Empty(t, m1.Keys())
		xt.Empty(t, m1.Map(false))
		xt.Empty(t, m1.Map(true))
	}
	check(t, &xmap.Ordered[int, string]{})
	check(t, &xmap.OrderedSync[int, string]{})
}
