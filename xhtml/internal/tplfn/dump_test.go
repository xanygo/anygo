//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-06

package tplfn

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestDump(t *testing.T) {
	data1 := map[any]any{
		"server":  "redis",
		"version": "8.2.1",
		"proto":   int64(3),
		"id":      int64(20),
		"mode":    "standalone",
		"role":    "master",
		"modules": []any{
			map[any]any{
				"name": "ReJSON",
				"ver":  int64(80200),
				"path": "",
				"args": []any{},
			},
			map[any]any{
				"name": "timeseries",
				"ver":  int64(80200),
				"path": "",
				"args": []any{},
			},
		},
	}
	code := Dump(data1)
	t.Logf("dump code:%s\n", code)
	xt.NotEmpty(t, code)
}
