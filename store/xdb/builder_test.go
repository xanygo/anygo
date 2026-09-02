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
	xt.Equal(t, str, "INSERT INTO user (id,name) VALUES (?,?)")
	xt.Equal(t, arg, []any{1, "hello"})
}

func TestCondition_Build(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		cond := xdb.Condition{}
		cond.And("a=?", 1)
		cond.And("b=?", 2)
		where, args := cond.MustBuild()
		xt.Equal(t, where, "a=? AND b=?")
		xt.Equal(t, args, []any{1, 2})

		cond.And("c=? or d=?", 3, 4)
		where, args = cond.MustBuild()
		xt.Equal(t, where, "a=? AND b=? AND (c=? or d=?)")
		xt.Equal(t, args, []any{1, 2, 3, 4})
	})
	t.Run("case 2", func(t *testing.T) {
		cond := xdb.Condition{}
		cond.And("a=?", 1)
		cond.And("b=?", 2)
		cond.Or("c=? and d=?", 3, 4)
		where, args := cond.MustBuild()
		xt.Equal(t, where, "a=? AND b=? OR (c=? and d=?)")
		xt.Equal(t, args, []any{1, 2, 3, 4})
	})
}
