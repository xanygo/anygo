//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestGet(t *testing.T) {
	var m1 map[string]int

	got1, ok1 := Get(m1, "k1")
	xt.False(t, ok1)
	xt.Empty(t, got1)
	xt.Equal(t, 0, GetDf(m1, "k1", 0))
	xt.Equal(t, 2, GetDf(m1, "k1", 2))

	m1 = map[string]int{"k1": 1}
	got2, ok2 := Get(m1, "k1")
	xt.True(t, ok2)
	xt.Equal(t, 1, got2)
	xt.Equal(t, 1, GetDf(m1, "k1", 0))
	xt.Equal(t, 1, GetDf(m1, "k1", 2))

	got3, ok3 := Get(m1, "k2")
	xt.False(t, ok3)
	xt.Equal(t, 0, got3)
	xt.Equal(t, 0, GetDf(m1, "k2", 0))
	xt.Equal(t, 2, GetDf(m1, "k2", 2))
}

func TestRange(t *testing.T) {
	t.Run("string key map", func(t *testing.T) {
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
		xt.SliceSortEqual(t, wantKeys, keys)
		xt.Equal(t, 4, num)

		keys = nil
		num = Range[string, int](mp, func(key string, val int) bool {
			keys = append(keys, key)
			return true
		})
		xt.SliceSortEqual(t, []string{"k1", "k2"}, keys)
		xt.Equal(t, 2, num)
	})

	t.Run("nil map", func(t *testing.T) {
		num := Range[string, any](nil, func(key string, val any) bool {
			return true
		})
		xt.Equal(t, 0, num)
	})

	t.Run("empty map", func(t *testing.T) {
		var m map[string]any
		num := Range[string, any](m, func(key string, val any) bool {
			return true
		})
		xt.Equal(t, 0, num)
	})
}
