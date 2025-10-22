//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestGet(t *testing.T) {
	var m1 map[string]int

	got1, ok1 := Get(m1, "k1")
	fst.False(t, ok1)
	fst.Empty(t, got1)
	fst.Equal(t, 0, GetDf(m1, "k1", 0))
	fst.Equal(t, 2, GetDf(m1, "k1", 2))

	m1 = map[string]int{"k1": 1}
	got2, ok2 := Get(m1, "k1")
	fst.True(t, ok2)
	fst.Equal(t, 1, got2)
	fst.Equal(t, 1, GetDf(m1, "k1", 0))
	fst.Equal(t, 1, GetDf(m1, "k1", 2))

	got3, ok3 := Get(m1, "k2")
	fst.False(t, ok3)
	fst.Equal(t, 0, got3)
	fst.Equal(t, 0, GetDf(m1, "k2", 0))
	fst.Equal(t, 2, GetDf(m1, "k2", 2))
}

func TestRange(t *testing.T) {
	mp := map[string]any{
		"k1": 1,
		"k2": 2,
		"k3": []string{"1"},
		"k4": map[string]string{"1": "2"},
	}
	var keys []string
	num := Range[string, any](mp, func(key string, val any) bool {
		keys = append(keys, key)
		return true
	})
	wantKeys := Keys(mp)
	fst.SliceSortEqual(t, wantKeys, keys)
	fst.Equal(t, 4, num)

	keys = nil
	num = Range[string, int](mp, func(key string, val int) bool {
		keys = append(keys, key)
		return true
	})
	fst.SliceSortEqual(t, []string{"k1", "k2"}, keys)
	fst.Equal(t, 2, num)

	num = Range[string, any](nil, func(key string, val any) bool {
		return true
	})
	fst.Equal(t, 0, num)

	var m map[string]any
	num = Range[string, any](m, func(key string, val any) bool {
		return true
	})
	fst.Equal(t, 0, num)
}
