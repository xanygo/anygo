package buildin

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xvld/internal"
)

func NewTestContext1(current reflect.Value) *FieldContext {
	field := reflect.StructField{Name: "f1"}
	return internal.NewFieldContext(field, reflect.Value{}, reflect.Value{}, current, func(name string, params ...any) error {
		e, ok := Errs[name]
		if ok {
			return e.WithParams(params...)
		}
		return internal.ErrDefault.WithParams(params...)
	})
}

func TestRequired(t *testing.T) {
	rule := Rule{Name: "required"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		"",
		[]int{}, []string{},
		false,
		Rule{},
		map[string]string{},
		(map[string]string)(nil),
		[0]int{},
		nil,
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := required(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		int(1), int8(1), int16(-1), int32(2), int64(3),
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5),
		float32(1), float64(2),
		"hello",
		[]int{1}, []string{"abc"}, []string{""},
		true,
		rule,
		map[string]string{"a": "v"},
		[1]int{},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := required(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}
func TestOneof(t *testing.T) {
	rule := Rule{Name: "oneof", Param: "1|2|3"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		false, true,
		Rule{}, rule,
		nil,
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := oneOf(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		// 下面的特殊值，会忽略为空的情况
		"", []int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},

		int(1), int8(1), int16(1), int32(2), int64(3),
		uint(1), uint8(2), uint16(3), uint32(2), uint64(1),
		float32(1), float64(2),
		"2",
		[]int{1}, []string{"1"}, []string{"2"}, []string{"1", "2"},
		map[string]string{"a": "1"},
		[1]int{1},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := oneOf(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}

func TestIsURL(t *testing.T) {
	rule := Rule{Name: "url"}
	xt.True(t, isURL("http://www.example.com/", rule))
	xt.True(t, isURL("https://www.example.com/", rule))

	xt.False(t, isURL("ftp://www.example.com/", rule))
	xt.False(t, isURL("www.example.com/", rule))
}

func TestNotZeroUint(t *testing.T) {
	rule := Rule{Name: "nz-uint"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		int16(-1),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		float32(1), float64(2), // 不支持 float
		"hello",
		false, true,
		Rule{},
		nil,
		rule,
		map[string]string{"a": "v"},
		[1]int{},
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := notZeroUint(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		// 特殊-空置
		[]int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},

		int(1), int8(1), int16(1), int32(2), int64(3),
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5),

		[]int{1},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := notZeroUint(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}
