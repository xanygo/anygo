//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xkv_test

import (
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/store/xkv"
)

func TestFileStorage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "xkv_file")
	ff := &xkv.FileStore{
		DataDir: dir,
	}
	testStorage(t, ff)
}
