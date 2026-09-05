//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-24

package xkv_test

import (
	"testing"

	"github.com/xanygo/anygo/store/xkv"
)

func TestMemory(t *testing.T) {
	ff := &xkv.Memory{}
	testStringStorage(t, ff)
}

func BenchmarkMemory(b *testing.B) {
	ff := &xkv.Memory{}
	benchStorage(b, ff)
}
