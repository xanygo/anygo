package buildin

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xvld/internal"
)

func NewTestContext1(current reflect.Value) *FieldContext {
	field := reflect.StructField{Name: "f1"}
	newErr := func(name string, params ...any) error {
		e, ok := Errs[name]
		if ok {
			return e.WithParams(params...)
		}
		return fmt.Errorf("err=%q %w", name, xerror.NotFound)
	}
	return internal.NewFieldContext(field, reflect.Value{}, reflect.Value{}, current, newErr)
}

func Test_required(t *testing.T) {
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
		time.Time{},
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
		time.Now(),
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := required(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}

func Test_oneof(t *testing.T) {
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

func Test_isURL(t *testing.T) {
	rule := Rule{Name: "url"}
	xt.True(t, isURL("http://www.example.com/", rule))
	xt.True(t, isURL("https://www.example.com/", rule))

	xt.False(t, isURL("ftp://www.example.com/", rule))
	xt.False(t, isURL("www.example.com/", rule))
}

func Test_notZeroUint(t *testing.T) {
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

func Test_validatePort(t *testing.T) {
	rule := Rule{Name: "port"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		int16(-1),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		float32(1), float64(2), // 不支持 float
		70000,
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
			err1 := validatePort(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		// 特殊-空置
		[]int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},

		int(1), int8(1), int16(1), int32(2), int64(3),
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5),
		[]int{1},
		"80", 443,
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validatePort(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}

func Test_validateIP(t *testing.T) {
	rule := Rule{Name: "ip"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		int16(-1),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		float32(1), float64(2),
		int(1), int8(1), int16(1), int32(2), int64(3),
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5),
		[]int{1},
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
			err1 := validateIP(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		// 特殊-空置
		[]int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},
		"127.0.0.1", "::1", "10.0.0.1",
		"0.0.0.0", "127.0.0.0",
		[]byte("127.0.0.1"),
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateIP(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}

func Test_validateIPV4(t *testing.T) {
	rule := Rule{Name: "ipv4"}
	falseValues := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		int16(-1),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0),
		float32(1), float64(2),
		int(1), int8(1), int16(1), int32(2), int64(3),
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5),
		[]int{1},
		"hello",
		false, true,
		Rule{},
		nil,
		rule,
		map[string]string{"a": "v"},
		[1]int{},
		"::1", "2001:0DB8:02de::0e13",
	}
	for i, v := range falseValues {
		t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateIPV4(fc1, rule)
			xt.Error(t, err1, "value=%#v", v)
		})
	}

	trueValues := []any{
		// 特殊-空置
		[]int{}, []string{}, map[string]string{}, (map[string]string)(nil), [0]int{},
		"127.0.0.1", "10.0.0.1", "0.0.0.0", "127.0.0.0",
		[]byte("127.0.0.1"), net.IP("127.0.0.1"),
	}
	for i, v := range trueValues {
		t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
			fc1 := NewTestContext1(reflect.ValueOf(v))
			err1 := validateIPV4(fc1, rule)
			xt.NoError(t, err1, "value=%#v", v)
		})
	}
}
