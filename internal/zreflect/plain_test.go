package zreflect_test

import (
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
	"testing"
)

func TestToPlainObject(t *testing.T) {
	xt.Equal(t, zreflect.ToPlainObject(nil), nil)
	xt.Equal(t, zreflect.ToPlainObject(123), 123)

	xt.Equal[any](t, zreflect.ToPlainObject([]string{"hello"}), []string{"hello"})

	p1 := Person{
		Name: "John Doe",
		Age:  18,
	}
	pw1 := map[string]any{
		"Name": "John Doe",
		"Age":  18,
	}
	xt.Equal[any](t, zreflect.ToPlainObject(p1), pw1)
	xt.Equal[any](t, zreflect.ToPlainObject(&p1), pw1)
}

type Person struct {
	Name string `json:"name"`
	Age  int
}
