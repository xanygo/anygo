//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xcache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xcodec"
)

func TestFile(t *testing.T) {
	c1 := &File[string, int]{
		Dir:   filepath.Join("tmp", "file"),
		Codec: xcodec.JSON,
	}
	testCache(t, c1)

	os.RemoveAll("tmp")
}
