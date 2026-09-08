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
		err := Range[string, any](mp, func(key string, val any) bool {
			keys = append(keys, key)
			return true
		})
		xt.NoError(t, err)
		wantKeys := Keys(mp)
		xt.SliceSortEqual(t, wantKeys, keys)

		keys = nil
		err = Range[string, int](mp, func(key string, val int) bool {
			keys = append(keys, key)
			return true
		})
		xt.SliceSortEqual(t, []string{"k1", "k2"}, keys)
		xt.NoError(t, err)
	})

	t.Run("nil", func(t *testing.T) {
		err := Range[string, any](nil, func(key string, val any) bool {
			return true
		})
		xt.Error(t, err)
	})

	t.Run("empty map", func(t *testing.T) {
		var m map[string]any
		err := Range[string, any](m, func(key string, val any) bool {
			return true
		})
		xt.NoError(t, err)
	})
}

func TestGetString(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var data map[string]any
		got, err := GetString(data, "k")
		xt.NoError(t, err)
		xt.Empty(t, got)
	})
	t.Run("case 2", func(t *testing.T) {
		data := map[string]any{"k1": "123", "k2": 234}
		got, err := GetString(data, "k1")
		xt.NoError(t, err)
		xt.Equal(t, got, "123")

		got, err = GetString(data, "k2")
		xt.NoError(t, err)
		xt.Equal(t, got, "234")
	})
}

func TestGetMap(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var data map[string]any
		got, err := GetMap(data, "k")
		xt.NoError(t, err)
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
		got, err := GetMap(data, "k1")
		xt.Error(t, err)
		xt.Empty(t, got)

		got, err = GetMap(data, "k3")
		xt.NoError(t, err)
		xt.Equal(t, got, map[string]any{"t1": "v2"})

		got, err = GetMap(data, "k4")
		xt.NoError(t, err)
		xt.Equal(t, got, map[string]any{"t1": "v2"})

		got, err = GetMap(data, "k5")
		xt.NoError(t, err)
		xt.Equal(t, got, map[string]any{"t1": "v2"})
	})
}
