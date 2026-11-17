package buildin

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func Test_validateContains_str(t *testing.T) {
	rule1 := Rule{Name: "contains", Param: "a"}
	rule2 := Rule{Name: "contains", Param: "a|b"}
	falseValues := []any{
		int(0),
		uint(0),
		float32(0),
		[]int{1},
		"hello",
		false, true,
		Rule{},
		nil,
		rule1,
		map[string]string{"a": "v"},
		map[string]string{}, (map[string]string)(nil),
		[1]int{},
	}

	trueValues := []any{
		// 特殊-空置
		[]int{}, []string{}, [0]int{},
		"abc", []byte("h-a-b"),
	}

	t.Run("rule1", func(t *testing.T) {
		for i, v := range falseValues {
			t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule1)
				xt.Error(t, err1, "value=%#v", v)
			})
		}

		for i, v := range trueValues {
			t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule1)
				xt.NoError(t, err1, "value=%#v", v)
			})
		}
	})
	t.Run("rule2", func(t *testing.T) {
		for i, v := range falseValues {
			t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule2)
				xt.Error(t, err1, "value=%#v", v)
			})
		}

		for i, v := range trueValues {
			t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule2)
				xt.NoError(t, err1, "value=%#v", v)
			})
		}
	})
}
func Test_validateContains_int(t *testing.T) {
	rule1 := Rule{Name: "contains", Param: "1"}
	rule2 := Rule{Name: "contains", Param: "1|2"}
	falseValues := []any{
		[]int{0},
		[]int{3},
		[]uint{3},
		[]float32{3},
	}

	trueValues := []any{
		[]int{0, 1},
		[]int{0, 2, 1},
		[]uint{10, 1, 3},
		[]float32{10, 2, 0, 1},
	}

	t.Run("rule1", func(t *testing.T) {
		for i, v := range falseValues {
			t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule1)
				xt.Error(t, err1, "value=%#v", v)
			})
		}

		for i, v := range trueValues {
			t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule1)
				xt.NoError(t, err1, "value=%#v", v)
			})
		}
	})
	t.Run("rule2", func(t *testing.T) {
		for i, v := range falseValues {
			t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule2)
				xt.Error(t, err1, "value=%#v", v)
			})
		}

		for i, v := range trueValues {
			t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateContains(fc1, rule2)
				xt.NoError(t, err1, "value=%#v", v)
			})
		}
	})
}

func Test_validateExcludes_int(t *testing.T) {
	rule1 := Rule{Name: "excludes", Param: "1"}
	rule2 := Rule{Name: "excludes", Param: "1|2"}
	falseValues := []any{
		[]int{0, 1},
		[]int{0, 2, 1},
		[]uint{10, 1, 3},
		[]float32{10, 2, 0, 1},
	}

	trueValues := []any{
		[]int{0},
		[]int{3},
		[]uint{3},
		[]float32{3},
	}

	t.Run("rule1", func(t *testing.T) {
		for i, v := range falseValues {
			t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateExcludes(fc1, rule1)
				xt.Error(t, err1, "value=%#v", v)
			})
		}

		for i, v := range trueValues {
			t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateExcludes(fc1, rule1)
				xt.NoError(t, err1, "value=%#v", v)
			})
		}
	})
	t.Run("rule2", func(t *testing.T) {
		for i, v := range falseValues {
			t.Run(fmt.Sprintf("false-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateExcludes(fc1, rule2)
				xt.Error(t, err1, "value=%#v", v)
			})
		}

		for i, v := range trueValues {
			t.Run(fmt.Sprintf("true-%d", i), func(t *testing.T) {
				fc1 := NewTestContext1(reflect.ValueOf(v))
				err1 := validateExcludes(fc1, rule2)
				xt.NoError(t, err1, "value=%#v", v)
			})
		}
	})
}
