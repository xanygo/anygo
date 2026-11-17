package buildin

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func Test_minMax_min(t *testing.T) {
	rule := Rule{Name: "min", Param: "10"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		false, true,
		Rule{}, rule,
		nil,
		map[string]any{"a": "1", "b": 1},
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := minMax(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		int(10), int8(11), int16(10), int32(20), int64(30),
		uint(11), uint8(20), uint16(30), uint32(40), uint64(50),
		float32(10), float64(20),
		"", []int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{}, // 空值不会校验
		map[string]string{},
		"hello",
		[]int{10}, []string{"abc"}, []string{"b"},
		map[string]any{"a": "10", "b": 12},
		[1]int{10},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := minMax(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}
func Test_minMax_max(t *testing.T) {
	rule := Rule{Name: "max", Param: "20"}
	falseValues := []any{
		int(100), int8(100), int16(1000), int32(2000), int64(200),
		uint(21), uint8(22), uint16(23), uint32(24), uint64(25),
		float32(26), float64(28),
		false, true,
		Rule{}, rule,
		nil,
		map[string]any{"a": "21", "b": 21},
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := minMax(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		int(10), int8(11), int16(10), int32(20), int64(20),
		uint(11), uint8(20), uint16(10), uint32(19), uint64(10),
		float32(10), float64(20),
		"", []int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{}, // 空值不会校验
		map[string]string{},
		"10",
		[]int{10}, []string{"1"}, []string{"0"},
		map[string]any{"a": "10", "b": 12},
		[1]int{10},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := minMax(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}

func Test_validateLen_min(t *testing.T) {
	rule := Rule{Name: "min-len", Param: "2"}
	falseValues := []any{
		int(100), int8(100), int16(1000), int32(2000), int64(200),
		uint(21), uint8(22), uint16(23), uint32(24), uint64(25),
		float32(26), float64(28),
		false, true,
		Rule{}, rule,
		nil,
		"", []int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},
		[]string{"0"},
		map[string]any{"a": "21"},
		[1]int{10},
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateLen(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		map[string]string{"a": "1", "b": "2"},
		"10",
		[]int{10, 20}, []string{"1", "2"},
		map[string]any{"a": "10", "b": 12},
		[2]int{10, 20},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateLen(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}
func Test_validateLen_max(t *testing.T) {
	rule := Rule{Name: "max-len", Param: "2"}
	falseValues := []any{
		int(100), int8(100), int16(1000), int32(2000), int64(200),
		uint(21), uint8(22), uint16(23), uint32(24), uint64(25),
		float32(26), float64(28),
		false, true,
		Rule{}, rule,
		nil,
		map[string]any{"a": "21", "b": "", "c": ""},
		"101",
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateLen(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		map[string]string{"a": "1", "b": "2"},
		[]int{10, 20}, []string{"1", "2"},
		[]string{"0"},
		"", []int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},
		map[string]any{"a": "10", "b": 12},
		[2]int{10, 20},
		[1]int{10},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateLen(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}

func Test_validateNotEqual(t *testing.T) {
	rule := Rule{Name: "ne", Param: "2"}
	falseValues := []any{
		"2", 2, uint(2), []int{2}, []string{"2"},
		map[string]any{"a": "2", "b": 1}, map[string]any{"a": "3", "b": 2},
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateNotEqual(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		"1", 3, uint(4), []int{3}, []string{"6"},
		map[string]any{"a": "3", "b": 4}, map[string]any{"a": "3", "b": 4},
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateNotEqual(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}
