//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-9-5

package xkvx

import (
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestConfigFileLoad(t *testing.T) {
	cf := &ConfigFile{
		Items: []map[string]any{
			{
				"Name": "kv1",
				"Type": "File",
				"Dir":  filepath.Join(t.TempDir(), "kv1"),
			},
			{
				"Name": "kv2",
				"Type": "Memory",
			},
			{
				"Name": "kv3",
				"Type": "Nop",
			},
		},
	}
	t.Run("string", func(t *testing.T) {
		kv1, err := cf.Load[string]("kv1")
		xt.NoError(t, err)
		xt.NotNil(t, kv1)

		kv2, err := cf.Load[string]("kv2")
		xt.NoError(t, err)
		xt.NotNil(t, kv2)

		kv3, err := cf.Load[string]("kv3")
		xt.NoError(t, err)
		xt.NotNil(t, kv3)
	})
	t.Run("int", func(t *testing.T) {
		kv1, err := cf.Load[int]("kv1")
		xt.NoError(t, err)
		xt.NotNil(t, kv1)

		kv2, err := cf.Load[int]("kv2")
		xt.NoError(t, err)
		xt.NotNil(t, kv2)

		kv3, err := cf.Load[int]("kv3")
		xt.NoError(t, err)
		xt.NotNil(t, kv3)
	})
}
