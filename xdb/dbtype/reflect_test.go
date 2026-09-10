package dbtype

import (
	"reflect"
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestReflectToKind(t *testing.T) {
	t.Run("time", func(t *testing.T) {
		got, ok := ReflectToKind(reflect.TypeOf(time.Now()))
		xt.True(t, ok)
		xt.Equal(t, got, KindDateTime)
	})
	t.Run("bytes", func(t *testing.T) {
		var b []byte
		got, ok := ReflectToKind(reflect.TypeOf(b))
		xt.True(t, ok)
		xt.Equal(t, got, KindBinary)
	})

	t.Run("array_bytes", func(t *testing.T) {
		var b [1]byte
		rt := reflect.TypeOf(b)
		got, ok := ReflectToKind(rt)
		xt.True(t, ok)
		xt.Equal(t, got, KindBinary)
	})
}
