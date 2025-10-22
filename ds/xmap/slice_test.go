//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-01

package xmap_test

import (
	"iter"
	"sort"
	"testing"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xt"
)

type testSliceValueSyncType[K, V comparable] interface {
	Set(key K, values ...V)
	Get(key K) []V
	GetFirst(key K) (v V)
	AddUnique(key K, values ...V)
	Delete(keys ...K)
	DeleteValue(key K, values ...V)
	Has(key K) bool
	HasValue(key K, values ...V) bool
	Keys() []K
	Map(clone bool) map[K][]V
	Iter() iter.Seq2[K, []V]
}

var (
	_ testSliceValueSyncType[int, int] = (*xmap.SliceValue[int, int])(nil)
	_ testSliceValueSyncType[int, int] = (*xmap.SliceValueSync[int, int])(nil)
)

func TestSlice(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		check := func(t *testing.T, s1 testSliceValueSyncType[string, int]) {
			s1.Set("k1", 1)
			xt.Equal(t, []int{1}, s1.Get("k1"))
			xt.Equal(t, 1, s1.GetFirst("k1"))
			xt.True(t, s1.Has("k1"))
			xt.False(t, s1.Has("k2"))
			xt.True(t, s1.HasValue("k1", 1, 2))
			xt.False(t, s1.HasValue("k2", 1, 2))

			s1.Set("k1", 2)
			xt.Equal(t, []int{2}, s1.Get("k1"))

			s1.AddUnique("k2", 2, 3)
			xt.True(t, s1.Has("k2"))
			xt.True(t, s1.HasValue("k2", 2, 3))
			xt.False(t, s1.HasValue("k2", 4))

			s1.AddUnique("k2", 4)
			xt.Equal(t, []int{2, 3, 4}, s1.Get("k2"))

			keys := s1.Keys()
			sort.Strings(keys)
			xt.Equal(t, []string{"k1", "k2"}, keys)
			xt.Equal(t, map[string][]int{"k1": {2}, "k2": {2, 3, 4}}, s1.Map(false))

			s1.Delete("k3", "k1")
			s1.DeleteValue("k2", 3, 1)
			xt.Equal(t, map[string][]int{"k2": {2, 4}}, s1.Map(true))
		}
		check(t, &xmap.SliceValue[string, int]{})
		check(t, &xmap.SliceValueSync[string, int]{})
	})

	t.Run("case 2", func(t *testing.T) {
		s1 := &xmap.SliceValueSync[string, int]{}
		s1.Set("k1", 1)
		xt.Equal(t, []int{1}, s1.Get("k1"))
	})
}
