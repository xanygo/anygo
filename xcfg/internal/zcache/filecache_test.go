//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package zcache

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestFileCache_Get(t *testing.T) {
	fc := &FileCache{
		Dir: filepath.Join("testdata", "cache"),
	}
	val1 := []byte("hello")
	fc.Set("k1", val1)
	got1, ok1 := fc.Get("k1", 0)
	fst.True(t, ok1)
	fst.Equal(t, string(val1), string(got1))

	got1, ok1 = fc.Get("k1", time.Hour)
	fst.True(t, ok1)
	fst.Equal(t, string(val1), string(got1))
}
