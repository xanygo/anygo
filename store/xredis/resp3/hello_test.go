//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-06

package resp3

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/fsgo/fst"

	"github.com/xanygo/anygo/xcodec/xbase"
)

func TestHelloResponse(t *testing.T) {
	content, err := xbase.ReadBase64File("testdata/hello_resp3.b64")
	fst.NoError(t, err)

	mp, err1 := ReadByType(bufio.NewReader(bytes.NewBuffer(content)), DataTypeMap)
	fst.NoError(t, err1)
	fst.NotEmpty(t, mp)
	fst.Equal(t, DataTypeMap, mp.DataType())
	mv, err2 := ToAny(mp)
	fst.NoError(t, err2)
	fst.NotEmpty(t, mv)
	obj, ok1 := mv.(map[any]any)
	fst.True(t, ok1)

	want := map[any]any{
		"server":  "redis",
		"version": "8.2.1",
		"proto":   int64(3),
		"id":      int64(21),
		"mode":    "standalone",
		"role":    "master",
		"modules": []any{
			map[any]any{
				"name": "timeseries",
				"ver":  int64(80200),
				"path": "/usr/lib/redis/modules/redistimeseries.so",
				"args": []any{},
			},
			map[any]any{
				"name": "ReJSON",
				"ver":  int64(80200),
				"path": "/usr/lib/redis/modules/rejson.so",
				"args": []any{},
			},
			map[any]any{
				"name": "vectorset",
				"ver":  int64(1),
				"path": "",
				"args": []any{},
			},
			map[any]any{
				"name": "bf",
				"ver":  int64(80200),
				"path": "/usr/lib/redis/modules/redisbloom.so",
				"args": []any{},
			},
			map[any]any{
				"name": "search",
				"ver":  int64(80201),
				"path": "/usr/lib/redis/modules/redisearch.so",
				"args": []any{},
			},
		},
	}
	fst.Equal(t, want, obj)
}
