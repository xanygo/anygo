//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-03

package xfs

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestExists(t *testing.T) {
	ok1, err1 := Exists("helper.go")
	xt.True(t, ok1)
	xt.NoError(t, err1)

	ok2, err2 := Exists("file_not.go")
	xt.False(t, ok2)
	xt.NoError(t, err2)

	ok3, err3 := Exists("testdata")
	xt.True(t, ok3)
	xt.NoError(t, err3)
}
