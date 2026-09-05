package zreflect_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestTypeID1(t *testing.T) {
	xt.Equal(t, zreflect.TypeID1[string](), 1)
	xt.Equal(t, zreflect.TypeID1[[]byte](), 15)

	xt.Equal(t, zreflect.TypeID("string"), 1)
	xt.Equal(t, zreflect.TypeID([]byte("hello")), 15)
}

func TestTypeID2(t *testing.T) {
	xt.Equal(t, zreflect.TypeID2[string, string](), 1)
	xt.Equal(t, zreflect.TypeID("a", "b"), 1)

	xt.Equal(t, zreflect.TypeID2[[]byte, []byte](), 15)
	xt.Equal(t, zreflect.TypeID([]byte("a"), []byte("b")), 15)
}

func TestTypeID(t *testing.T) {
	xt.Equal(t, zreflect.TypeID(), 0)
	xt.Equal(t, zreflect.TypeID(nil), 2359885923)
	xt.Equal(t, zreflect.TypeID("string", 123), 2289370667)
}
