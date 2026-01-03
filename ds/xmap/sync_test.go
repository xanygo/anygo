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
		xt.Equal(t, "", v1)
		xt.False(t, ok1)

		m1.Store(key1, "world")
		xt.Equal(t, map[string]string{"hello": "world"}, m1.ToMap())
		xt.Equal(t, 1, m1.Len())

		v1, ok1 = m1.Load(key1)
		xt.Equal(t, "world", v1)
		xt.True(t, ok1)

		m1.Delete(key1)

		v1, ok1 = m1.Load(key1)
		xt.Equal(t, "", v1)
		xt.False(t, ok1)

		v2, ok2 := m1.LoadOrStore(key1, "h1")
		xt.Equal(t, "h1", v2)
		xt.False(t, ok2)

		v2, ok2 = m1.LoadOrStore(key1, "h2")
		xt.Equal(t, "h1", v2)
		xt.True(t, ok2)

		var num1 int
		m1.Range(func(key string, value string) bool {
			num1++
			xt.Equal(t, key1, key)
			xt.Equal(t, "h1", value)
			return true
		})

		xt.Equal(t, 1, num1)

		v3, ok3 := m1.LoadAndDelete(key1)
		xt.Equal(t, "h1", v3)
		xt.True(t, ok3)

		v3, ok3 = m1.LoadAndDelete(key1)
		xt.Equal(t, "", v3)
		xt.False(t, ok3)
	})

	t.Run("case 2", func(t *testing.T) {
		var m2 Sync[string, error]
		got, ok := m2.LoadOrStore("k1", io.EOF)
		xt.False(t, ok)
		xt.Equal(t, io.EOF, got)
	})

	t.Run("case 3", func(t *testing.T) {
		var m2 Sync[string, error]
		got, ok := m2.LoadAndDelete("k1")
		xt.False(t, ok)
		xt.Nil(t, got)
	})
}
