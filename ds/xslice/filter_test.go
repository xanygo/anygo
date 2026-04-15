//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-15

package xslice_test

import (
	"testing"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/xt"
)

type user struct {
	Name string
	Tags []string
}

var usersNeedFilter = []user{
	{
		Name: "name1",
		Tags: []string{"tag1", "tag2"},
	},
	{
		Name: "name1",
		Tags: []string{"tag1"},
	},
	{
		Name: "name3",
		Tags: []string{"tag2"},
	},
	{
		Name: "name4",
		Tags: []string{"tag3"},
	},
}

func TestBuildTagFirst(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		fn, err := xslice.BuildTagFirst("tag1,tag2,[ANY]", func(t user) []string {
			return t.Tags
		})
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		xt.Equal(t, usersNeedFilter[0], got)
	})
	t.Run("case 2", func(t *testing.T) {
		fn, err := xslice.BuildTagFirst("tag2,[ANY]", func(t user) []string {
			return t.Tags
		})
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		xt.Equal(t, usersNeedFilter[0], got)
	})
	t.Run("case 3", func(t *testing.T) {
		fn, err := xslice.BuildTagFirst("tag3,[ANY]", func(t user) []string {
			return t.Tags
		})
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		xt.Equal(t, usersNeedFilter[3], got)
	})
}

func TestBuildTagFilter(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		fn, err := xslice.BuildTagFilter[user]("tag1,tag2,[ANY]", func(t user) []string {
			return t.Tags
		}, 0)
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		want := []user{
			usersNeedFilter[0],
			usersNeedFilter[1],
			usersNeedFilter[2],
		}
		xt.Equal(t, want, got)
	})
	t.Run("case 2", func(t *testing.T) {
		fn, err := xslice.BuildTagFilter("tag2,[ANY]", func(t user) []string {
			return t.Tags
		}, 0)
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		want := []user{
			usersNeedFilter[0],
			usersNeedFilter[2],
		}
		xt.Equal(t, want, got)
	})
	t.Run("case 3", func(t *testing.T) {
		fn, err := xslice.BuildTagFilter("tag3,[ANY]", func(t user) []string {
			return t.Tags
		}, 0)
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		want := []user{
			usersNeedFilter[3],
		}
		xt.Equal(t, want, got)
	})

	t.Run("case 4", func(t *testing.T) {
		fn, err := xslice.BuildTagFilter("tag-not-found,[ANY]", func(t user) []string {
			return t.Tags
		}, 0)
		xt.NoError(t, err)
		got := fn(usersNeedFilter)
		xt.Len(t, got, 1)
	})
}
