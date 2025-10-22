//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import (
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestDefault(t *testing.T) {
	doInit()

	xt.NotEmpty(t, AppName())
	root := RootDir()
	xt.NotEmpty(t, root)
	xt.Equal(t, filepath.Join(root, "conf"), ConfDir())
	xt.Equal(t, filepath.Join(root, "data"), DataDir())
	xt.Equal(t, filepath.Join(root, "log"), LogDir())
	xt.Equal(t, filepath.Join(root, "temp"), TempDir())
	xt.Equal(t, IDCOnline, IDC())
	xt.Equal(t, ModeProduct, RunMode())

	Set("k1", "v1")
	got1, ok1 := Get("k1")
	xt.Equal(t, "v1", got1)
	xt.True(t, ok1)

	SetConfDir("/user/cfg|abs")
	xt.Equal(t, "/user/cfg", ConfDir())

	SetConfDir("user/cfg")
	xt.Equal(t, filepath.Join(root, "/user/cfg"), ConfDir())

	SetDataDir("/user/data|abs")
	xt.Equal(t, "/user/data", DataDir())

	SetDataDir("user/data")
	xt.Equal(t, filepath.Join(root, "/user/data"), DataDir())

	SetTempDir("/temp|abs")
	xt.Equal(t, "/temp", TempDir())

	SetTempDir("temp")
	xt.Equal(t, filepath.Join(root, "temp"), TempDir())

	SetLogDir("/temp/log|abs")
	xt.Equal(t, "/temp/log", LogDir())

	SetLogDir("temp/log")
	xt.Equal(t, filepath.Join(root, "temp", "log"), LogDir())
}
