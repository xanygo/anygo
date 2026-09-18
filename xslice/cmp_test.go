package xslice_test

import (
	"testing"

	"github.com/xanygo/anygo/xslice"
	"github.com/xanygo/anygo/xt"
)

func TestDiffMore(t *testing.T) {
	got := xslice.Difference([]string{"a", "b"}, nil)
	xt.Equal(t, got, []string{"a", "b"})

	got = xslice.Difference([]string{"a", "b"}, []string{"a"})
	xt.Equal(t, got, []string{"b"})
}
