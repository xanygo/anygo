//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-01

package xmap_test

import (
	"iter"
	"sort"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xmap"
)

type testSliceValueSyncType[K, V comparable] interface {
	Set(key K, values ...V)
	Get(key K) []V
	GetFirst(key K) (v V)
	Add(key K, values ...V)
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
			fst.Equal(t, []int{1}, s1.Get("k1"))
			fst.Equal(t, 1, s1.GetFirst("k1"))
			fst.True(t, s1.Has("k1"))
			fst.False(t, s1.Has("k2"))
			fst.True(t, s1.HasValue("k1", 1, 2))
			fst.False(t, s1.HasValue("k2", 1, 2))

			s1.Set("k1", 2)
			fst.Equal(t, []int{2}, s1.Get("k1"))

			s1.Add("k2", 2, 3)
			fst.True(t, s1.Has("k2"))
			fst.True(t, s1.HasValue("k2", 2, 3))
			fst.False(t, s1.HasValue("k2", 4))

			s1.Add("k2", 4)
			fst.Equal(t, []int{2, 3, 4}, s1.Get("k2"))

			keys := s1.Keys()
			sort.Strings(keys)
			fst.Equal(t, []string{"k1", "k2"}, keys)
			fst.Equal(t, map[string][]int{"k1": {2}, "k2": {2, 3, 4}}, s1.Map(false))

			s1.Delete("k3", "k1")
			s1.DeleteValue("k2", 3, 1)
			fst.Equal(t, map[string][]int{"k2": {2, 4}}, s1.Map(true))
		}
		check(t, &xmap.SliceValue[string, int]{})
		check(t, &xmap.SliceValueSync[string, int]{})
	})

	t.Run("case 2", func(t *testing.T) {
		s1 := &xmap.SliceValueSync[string, int]{}
		s1.Set("k1", 1)
		fst.Equal(t, []int{1}, s1.Get("k1"))
	})
}
