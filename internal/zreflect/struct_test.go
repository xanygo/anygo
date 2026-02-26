//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-13

package zreflect_test

import (
	"reflect"
	"testing"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestRangeStructFields(t *testing.T) {
	type structMeta struct {
		Fields []reflect.StructField
	}

	zreflect.StructMetaCache.Clear()
	t.Run("structMeta-ptr", func(t *testing.T) {
		t1 := reflect.TypeFor[*structMeta]()
		var names []string
		err := zreflect.RangeStructFields(t1, func(f reflect.StructField) error {
			names = append(names, f.Name)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, []string{"Fields"}, names)
		xt.Equal(t, 1, zreflect.StructMetaCache.Count())
	})

	t.Run("structMeta", func(t *testing.T) {
		t1 := reflect.TypeFor[structMeta]()
		var names []string
		err := zreflect.RangeStructFields(t1, func(f reflect.StructField) error {
			names = append(names, f.Name)
			return nil
		})
		xt.NoError(t, err)
		xt.Equal(t, []string{"Fields"}, names)

		xt.Equal(t, 2, zreflect.StructMetaCache.Count())
	})

	type user struct {
		Name  string
		age   int
		Class int
	}
	t.Run("user1", func(t *testing.T) {
		t1 := reflect.TypeFor[user]()
		var names []string
		err := zreflect.RangeStructFields(t1, func(f reflect.StructField) error {
			names = append(names, f.Name)
			return nil
		})
		xt.NoError(t, err)
		xt.SliceSortEqual(t, []string{"Name", "Class", "age"}, names)

		xt.Equal(t, 3, zreflect.StructMetaCache.Count())
	})
}

func BenchmarkRangeStructFields(b *testing.B) {
	zreflect.StructMetaCache.Clear()
	type user struct {
		Name  string
		age   int
		Class int
	}
	_ = user{Name: "a", age: 1, Class: 5}
	b.ResetTimer()
	b.Run("withCache", func(b *testing.B) {
		t := reflect.TypeFor[user]()
		for i := 0; i < b.N; i++ {
			zreflect.RangeStructFields(t, func(f reflect.StructField) error {
				return nil
			})
		}
	})
	b.Run("noCache", func(b *testing.B) {
		t := reflect.TypeFor[user]()
		var tmp reflect.StructField
		for i := 0; i < b.N; i++ {
			for field := range t.Fields() {
				tmp = field
			}
		}
		_ = tmp
	})
}
