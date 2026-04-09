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
	xt.Equal(t, ConfDir(), filepath.Join(root, "conf"))
	xt.Equal(t, DataDir(), filepath.Join(root, "data"))
	xt.Equal(t, LogDir(), filepath.Join(root, "log"))
	xt.Equal(t, TempDir(), filepath.Join(root, "temp"))
	xt.Equal(t, IDC(), IDCDev)
	xt.Equal(t, RunMode(), ModeDebug)

	Set("k1", "v1")
	got1, ok1 := Get("k1")
	xt.Equal(t, got1, "v1")
	xt.True(t, ok1)

	SetConfDir("/user/cfg|abs")
	xt.Equal(t, ConfDir(), "/user/cfg")

	SetConfDir("user/cfg")
	xt.Equal(t, ConfDir(), filepath.Join(root, "/user/cfg"))

	SetDataDir("/user/data|abs")
	xt.Equal(t, DataDir(), "/user/data")

	SetDataDir("user/data")
	xt.Equal(t, DataDir(), filepath.Join(root, "/user/data"))

	SetTempDir("/temp|abs")
	xt.Equal(t, TempDir(), "/temp")

	SetTempDir("temp")
	xt.Equal(t, TempDir(), filepath.Join(root, "temp"))

	SetLogDir("/temp/log|abs")
	xt.Equal(t, LogDir(), "/temp/log")

	SetLogDir("temp/log")
	xt.Equal(t, LogDir(), filepath.Join(root, "temp", "log"))
}

func TestGetAs(t *testing.T) {
	Set("TestGetAs-1", 123)
	got1, ok1 := Get("TestGetAs-1")
	xt.Equal(t, got1, 123)
	xt.True(t, ok1)

	got2, err2 := GetAs[int64]("TestGetAs-1")
	xt.Equal(t, got2, 123)
	xt.NoError(t, err2)
}
