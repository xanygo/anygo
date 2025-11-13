//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-13

package zreflect

import (
	"reflect"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestRangeStructFields(t *testing.T) {
	structMetaCache.Clear()
	t.Run("structMeta-ptr", func(t *testing.T) {
		m1 := &structMeta{}
		t1 := reflect.TypeOf(m1)
		var names []string
		err := RangeStructFields(t1, func(f reflect.StructField) error {
			names = append(names, f.Name)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, []string{"Fields"}, names)
		xt.Equal(t, 1, structMetaCache.Count())
	})

	t.Run("structMeta", func(t *testing.T) {
		m1 := structMeta{}
		t1 := reflect.TypeOf(m1)
		var names []string
		err := RangeStructFields(t1, func(f reflect.StructField) error {
			names = append(names, f.Name)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, []string{"Fields"}, names)

		xt.Equal(t, 2, structMetaCache.Count())
	})

	type user struct {
		Name  string
		age   int
		Class int
	}
	t.Run("user1", func(t *testing.T) {
		m1 := user{Name: "hello", age: 1, Class: 1}
		t1 := reflect.TypeOf(m1)
		var names []string
		err := RangeStructFields(t1, func(f reflect.StructField) error {
			names = append(names, f.Name)
			return nil
		})
		xt.NoError(t, err)
		xt.SliceSortEqual(t, []string{"Name", "Class", "age"}, names)

		xt.Equal(t, 3, structMetaCache.Count())
	})
}

func BenchmarkRangeStructFields(b *testing.B) {
	structMetaCache.Clear()
	type user struct {
		Name  string
		age   int
		Class int
	}
	_ = user{Name: "a", age: 1, Class: 5}
	b.ResetTimer()
	b.Run("withCache", func(b *testing.B) {
		u := user{}
		t := reflect.TypeOf(u)
		for i := 0; i < b.N; i++ {
			RangeStructFields(t, func(f reflect.StructField) error {
				return nil
			})
		}
	})
	b.Run("noCache", func(b *testing.B) {
		u := user{}
		t := reflect.TypeOf(u)
		var tmp reflect.StructField
		for i := 0; i < b.N; i++ {
			for z := 0; z < t.NumField(); z++ {
				tmp = t.Field(z)
			}
		}
		_ = tmp
	})
}
