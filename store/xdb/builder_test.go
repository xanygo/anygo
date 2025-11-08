//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-06

package xdb_test

import (
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"
)

func TestInsertBuilder_Build(t *testing.T) {
	t1 := xdb.NewInsertBuilder("user")
	str, arg, err := t1.Build()
	xt.Error(t, err)
	xt.Empty(t, str)
	xt.Empty(t, arg)
	t1.Values(map[string]any{"id": 1, "name": "hello"})
	str, arg, err = t1.Build()
	xt.NoError(t, err)
	xt.Equal(t, "INSERT INTO user (id,name) VALUES (?,?)", str)
	xt.Equal(t, []any{1, "hello"}, arg)
}
