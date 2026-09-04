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
	xt.Equal(t, GetDf(m1, "k1", 0), 0)
	xt.Equal(t, GetDf(m1, "k1", 2), 2)

	m1 = map[string]int{"k1": 1}
	got2, ok2 := Get(m1, "k1")
	xt.True(t, ok2)
	xt.Equal(t, got2, 1)
	xt.Equal(t, GetDf(m1, "k1", 0), 1)
	xt.Equal(t, GetDf(m1, "k1", 2), 1)

	got3, ok3 := Get(m1, "k2")
	xt.False(t, ok3)
	xt.Equal(t, got3, 0)
	xt.Equal(t, GetDf(m1, "k2", 0), 0)
	xt.Equal(t, GetDf(m1, "k2", 2), 2)
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
		ok := Range[string, any](mp, func(key string, val any) bool {
			keys = append(keys, key)
			return true
		})
		xt.True(t, ok)
		wantKeys := Keys(mp)
		xt.SliceSortEqual(t, wantKeys, keys)

		keys = nil
		ok = Range[string, int](mp, func(key string, val int) bool {
			keys = append(keys, key)
			return true
		})
		xt.SliceSortEqual(t, []string{"k1", "k2"}, keys)
		xt.True(t, ok)
	})

	t.Run("nil map", func(t *testing.T) {
		ok := Range[string, any](nil, func(key string, val any) bool {
			return true
		})
		xt.False(t, ok)
	})

	t.Run("empty map", func(t *testing.T) {
		var m map[string]any
		ok := Range[string, any](m, func(key string, val any) bool {
			return true
		})
		xt.True(t, ok)
	})
}

func TestGetString(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var data map[string]any
		got, ok := GetString(data, "k")
		xt.False(t, ok)
		xt.Empty(t, got)
	})
	t.Run("case 2", func(t *testing.T) {
		data := map[string]any{"k1": "123", "k2": 234}
		got, ok := GetString(data, "k1")
		xt.True(t, ok)
		xt.Equal(t, got, "123")

		got, ok = GetString(data, "k2")
		xt.True(t, ok)
		xt.Equal(t, got, "234")
	})
}

func TestGetMap(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var data map[string]any
		got, ok := GetMap(data, "k")
		xt.False(t, ok)
		xt.Empty(t, got)
	})

	t.Run("case 2", func(t *testing.T) {
		data := map[string]any{
			"k1": "123",
			"k2": 234,
			"k3": map[string]any{"t1": "v2"},
			"k4": map[any]any{"t1": "v2"},
			"k5": any(map[string]any{"t1": "v2"}),
		}
		got, ok := GetMap(data, "k1")
		xt.False(t, ok)
		xt.Empty(t, got)

		got, ok = GetMap(data, "k3")
		xt.True(t, ok)
		xt.Equal(t, got, map[string]any{"t1": "v2"})

		got, ok = GetMap(data, "k4")
		xt.True(t, ok)
		xt.Equal(t, got, map[string]any{"t1": "v2"})

		got, ok = GetMap(data, "k5")
		xt.True(t, ok)
		xt.Equal(t, got, map[string]any{"t1": "v2"})
	})
}
