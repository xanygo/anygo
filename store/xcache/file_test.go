//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xcodec"
)

func TestFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "xcache_file")
	c1 := &File[string, int]{
		Dir:   dir,
		Codec: xcodec.JSON,
	}
	testCache(t, c1)
}
