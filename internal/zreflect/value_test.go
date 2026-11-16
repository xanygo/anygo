package zreflect_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestIndirectValue(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		v1 := &http.Cookie{}
		v10 := zreflect.IndirectValue(reflect.ValueOf(v1))
		xt.True(t, v10.IsValid())
		xt.Equal(t, v10.Kind(), reflect.Struct)
		xt.Equal(t, v10.Type().Name(), "Cookie")

		xt.True(t, zreflect.IsEmptyValue(v10))

		v2 := &v1
		v20 := zreflect.IndirectValue(reflect.ValueOf(v2))
		xt.True(t, v20.IsValid())
		xt.Equal(t, v20.Kind(), reflect.Struct)
		xt.Equal(t, v20.Type().Name(), "Cookie")

		xt.True(t, zreflect.IsEmptyValue(v20))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 any
		v10 := zreflect.IndirectValue(reflect.ValueOf(v1))
		xt.False(t, v10.IsValid())
		xt.True(t, zreflect.IsEmptyValue(v10))

		v2 := &v1
		v20 := zreflect.IndirectValue(reflect.ValueOf(v2))
		xt.False(t, v20.IsValid())

		xt.True(t, zreflect.IsEmptyValue(v20))
	})

	t.Run("int-10", func(t *testing.T) {
		v1 := 10
		v10 := zreflect.IndirectValue(reflect.ValueOf(v1))
		xt.True(t, v10.IsValid())
		xt.False(t, zreflect.IsEmptyValue(v10))

		v2 := &v1
		v20 := zreflect.IndirectValue(reflect.ValueOf(v2))
		xt.True(t, v20.IsValid())

		xt.False(t, zreflect.IsEmptyValue(v20))
	})
	t.Run("int-0", func(t *testing.T) {
		v1 := 0
		v10 := zreflect.IndirectValue(reflect.ValueOf(v1))
		xt.True(t, v10.IsValid())
		xt.True(t, zreflect.IsEmptyValue(v10))

		v2 := &v1
		v20 := zreflect.IndirectValue(reflect.ValueOf(v2))
		xt.True(t, v20.IsValid())

		xt.True(t, zreflect.IsEmptyValue(v20))
	})
}

type value1 struct {
	P1 Person
	P2 *Person
}

type value2 struct {
	Person
	P1 Person
}

func TestRangeStructValue(t *testing.T) {
	t.Run("struct-Person", func(t *testing.T) {
		p1 := Person{
			Name: "hello",
			Age:  18,
		}
		var num int
		data := map[string]any{}
		zreflect.RangeStructValue(reflect.ValueOf(p1), func(field reflect.StructField, value reflect.Value) bool {
			num++
			switch field.Name {
			case "Name":
				data[field.Name] = value.String()
			case "Age":
				data[field.Name] = value.Int()
			}
			return true
		})
		xt.Equal(t, num, 2)
		xt.Equal(t, data, map[string]any{"Name": "hello", "Age": int64(18)})
	})

	t.Run("value1", func(t *testing.T) {
		v1 := value1{
			P1: Person{
				Name: "hello",
			},
			P2: &Person{
				Name: "world",
			},
		}
		var num int
		zreflect.RangeStructValue(reflect.ValueOf(v1), func(field reflect.StructField, value reflect.Value) bool {
			num++
			switch field.Name {
			case "P1":
				xt.True(t, zreflect.IsStructType(field.Type, false))
				xt.True(t, zreflect.IsStructType(field.Type, true))
			case "P2":
				xt.False(t, zreflect.IsStructType(field.Type, false))
				xt.True(t, zreflect.IsStructType(field.Type, true))
			default:
				t.Fatalf("unkonwn field %q", field.Name)
			}
			return true
		})
		xt.Equal(t, num, 2)
	})

	t.Run("value2", func(t *testing.T) {
		v1 := value2{
			Person: Person{
				Name: "hello",
			},
			P1: Person{
				Name: "world",
			},
		}
		var num int
		zreflect.RangeStructValue(reflect.ValueOf(v1), func(field reflect.StructField, value reflect.Value) bool {
			num++
			switch field.Name {
			case "Person":
				xt.True(t, zreflect.IsStructType(field.Type, false))
				xt.True(t, field.Anonymous)
			case "P1":
				xt.True(t, zreflect.IsStructType(field.Type, false))
				xt.False(t, field.Anonymous)
			default:
				t.Fatalf("unkonwn field %q", field.Name)
			}
			return true
		})
		xt.Equal(t, num, 2)
	})
}
