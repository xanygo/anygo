//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-24

package xkv_test

import (
	"testing"

	"github.com/xanygo/anygo/store/xkv"
)

func TestMemoryStorage(t *testing.T) {
	ff := &xkv.MemoryStore{}
	testStorage(t, ff)
}

func BenchmarkMemory(b *testing.B) {
	ff := &xkv.MemoryStore{}
	benchStorage(b, ff)
}
