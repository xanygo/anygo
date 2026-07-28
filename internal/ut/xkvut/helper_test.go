package xkvut_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/ut/xkvut"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

type testT struct {
	*testing.T
}

func (t testT) Run(name string, fn func(tb xt.TB)) {
	t.T.Run(name, func(tt *testing.T) {
		fn(testT{T: tt})
	})
}

func TestTestStringStorage1(t *testing.T) {
	kv := xkv.NewMemoryStore()
	xkvut.TestStringStorage1(testT{T: t}, kv)
}

func TestTestStringStorage2(t *testing.T) {
	kv := xkv.NewMemoryStore()
	xkvut.TestStringStorage2(testT{T: t}, kv)
}
