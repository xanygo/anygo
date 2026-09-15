//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-24

package xkvx_test

import (
	"testing"

	"github.com/xanygo/anygo/xkv/xkvx"
)

func TestMemory(t *testing.T) {
	ff := &xkvx.Memory{}
	testStringStorage(t, ff)
}

func BenchmarkMemory(b *testing.B) {
	ff := &xkvx.Memory{}
	benchStorage(b, ff)
}
