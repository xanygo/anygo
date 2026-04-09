//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xmap

import (
	"io"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestSync(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var m1 Sync[string, string]
		key1 := "hello"
		v1, ok1 := m1.Load(key1)
		xt.Equal(t, v1, "")
		xt.False(t, ok1)

		m1.Store(key1, "world")
		xt.Equal(t, m1.ToMap(), map[string]string{"hello": "world"})
		xt.Equal(t, m1.Len(), 1)

		v1, ok1 = m1.Load(key1)
		xt.Equal(t, v1, "world")
		xt.True(t, ok1)

		m1.Delete(key1)

		v1, ok1 = m1.Load(key1)
		xt.Equal(t, v1, "")
		xt.False(t, ok1)

		v2, ok2 := m1.LoadOrStore(key1, "h1")
		xt.Equal(t, v2, "h1")
		xt.False(t, ok2)

		v2, ok2 = m1.LoadOrStore(key1, "h2")
		xt.Equal(t, v2, "h1")
		xt.True(t, ok2)

		var num1 int
		m1.Range(func(key string, value string) bool {
			num1++
			xt.Equal(t, key, key1)
			xt.Equal(t, value, "h1")
			return true
		})

		xt.Equal(t, num1, 1)

		v3, ok3 := m1.LoadAndDelete(key1)
		xt.Equal(t, v3, "h1")
		xt.True(t, ok3)

		v3, ok3 = m1.LoadAndDelete(key1)
		xt.Equal(t, v3, "")
		xt.False(t, ok3)
	})

	t.Run("case 2", func(t *testing.T) {
		var m2 Sync[string, error]
		got, ok := m2.LoadOrStore("k1", io.EOF)
		xt.False(t, ok)
		xt.Equal(t, got, io.EOF)
	})

	t.Run("case 3", func(t *testing.T) {
		var m2 Sync[string, error]
		got, ok := m2.LoadAndDelete("k1")
		xt.False(t, ok)
		xt.Nil(t, got)
	})
}
