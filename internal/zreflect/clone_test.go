package zreflect_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestClone(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		now := time.Now()
		cp, err := zreflect.Clone(now)
		xt.NoError(t, err)
		xt.Equal(t, cp.String(), now.String())
	})

	t.Run("case 2", func(t *testing.T) {
		value := "abc"
		cp, err := zreflect.Clone(value)
		xt.NoError(t, err)
		xt.Equal(t, cp, value)
	})

	t.Run("case 3", func(t *testing.T) {
		value := map[string]any{"k1": "v1"}
		cp, err := zreflect.Clone(value)
		xt.NoError(t, err)
		xt.Equal(t, cp, value)
	})

	t.Run("case 4", func(t *testing.T) {
		value := Person{
			Name: "abc",
		}
		cp, err := zreflect.Clone(value)
		xt.NoError(t, err)
		xt.Equal(t, cp, value)
	})

	t.Run("case 5", func(t *testing.T) {
		value := &Person{
			Name: "abc",
		}
		cp, err := zreflect.Clone(value)
		xt.NoError(t, err)
		xt.Equal(t, cp, value)
	})

	t.Run("case 6", func(t *testing.T) {
		var value *Person
		cp, err := zreflect.Clone(value)
		xt.NoError(t, err)
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
		cp, err := zreflect.Clone(value)
		xt.NoError(t, err)
		xt.Equal(t, cp, value)
		xt.NotSamePtr(t, cp.P, value.P)
		xt.Equal(t, value.name, cp.name)
	})

	t.Run("case 8", func(t *testing.T) {
		u := url.Values{}
		u.Add("k1", "v1")
		cp, err := zreflect.Clone(u)
		xt.NoError(t, err)
		xt.Equal(t, u, cp)
		xt.Equal(t, u.Encode(), cp.Encode())
	})
}

type userClone struct {
	name string
	P    *Person
}
