package dbcodec

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestBinary_Encode(t *testing.T) {
	enc := Binary{}
	t.Run("case 1", func(t *testing.T) {
		a := []byte("hello")
		got, err := enc.Encode(a)
		xt.NoError(t, err)
		xt.Equal[any](t, got, a)
	})
	t.Run("case 2", func(t *testing.T) {
		a := [5]byte{'a', 'b', 'c', 'd', 'e'}
		got, err := enc.Encode(a)
		xt.NoError(t, err)
		xt.Equal[any](t, got, []byte("abcde"))
	})
	t.Run("case 3", func(t *testing.T) {
		a := 123
		got, err := enc.Encode(a)
		xt.Error(t, err)
		xt.Nil(t, got)
	})
}

func TestBinary_Decode(t *testing.T) {
	dec := Binary{}
	t.Run("case 1", func(t *testing.T) {
		var got []byte
		err := dec.Decode("hello", &got)
		xt.NoError(t, err)
		xt.Equal[any](t, got, []byte("hello"))
	})
	t.Run("case 2", func(t *testing.T) {
		var got [5]byte
		err := dec.Decode("hello", &got)
		xt.NoError(t, err)
		xt.Equal[any](t, got, [5]byte{'h', 'e', 'l', 'l', 'o'})
	})
	t.Run("case 3", func(t *testing.T) {
		var got [3]byte
		err := dec.Decode("hello", &got)
		xt.Error(t, err)
	})
}
