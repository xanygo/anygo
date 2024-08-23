//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-23

package fctime

import (
	"os"
	"runtime"
	"testing"

	"github.com/fsgo/fst"
)

func TestCtime(t *testing.T) {
	info, err := os.Stat("c_test.go")
	fst.NoError(t, err)
	fst.NotEmpty(t, info.Name())
	ctime := Ctime(info)
	t.Log(runtime.GOOS, runtime.GOARCH)
	t.Logf("%q ctime: %v", info.Name(), ctime)
	fst.NotEmpty(t, ctime)
}
