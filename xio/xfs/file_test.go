//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xfs

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestExists(t *testing.T) {
	ok1, err1 := Exists("file.go")
	fst.True(t, ok1)
	fst.NoError(t, err1)

	ok2, err2 := Exists("file_not.go")
	fst.False(t, ok2)
	fst.NoError(t, err2)

	ok3, err3 := Exists("testdata")
	fst.True(t, ok3)
	fst.NoError(t, err3)
}
