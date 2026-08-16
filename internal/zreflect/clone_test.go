package zreflect_test

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestClone(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		now := time.Now()
		cp := zreflect.Clone(now)
		xt.Equal(t, cp.String(), now.String())
	})

	t.Run("case 2", func(t *testing.T) {
		value := "abc"
		cp := zreflect.Clone(value)
		xt.Equal(t, cp, value)
	})

	t.Run("case 3", func(t *testing.T) {
		value := map[string]any{"k1": "v1"}
		cp := zreflect.Clone(value)
		xt.Equal(t, cp, value)
	})

	t.Run("case 4", func(t *testing.T) {
		value := Person{
			Name: "abc",
		}
		cp := zreflect.Clone(value)
		xt.Equal(t, cp, value)
	})
	t.Run("case 5", func(t *testing.T) {
		value := &Person{
			Name: "abc",
		}
		cp := zreflect.Clone(value)
		xt.Equal(t, cp, value)
	})
	t.Run("case 6", func(t *testing.T) {
		var value *Person
		cp := zreflect.Clone(value)
		xt.Equal(t, cp, value)
		xt.Nil(t, cp)
	})
	t.Run("case 7", func(t *testing.T) {
		value := &userClone{
			name: "hello",
			P: &Person{
				Name: "world",
			},
		}
		cp := zreflect.Clone(value)
		xt.Equal(t, cp, value)
		xt.NotSamePtr(t, cp.P, value.P)
	})
}

type userClone struct {
	name string
	P    *Person
}
