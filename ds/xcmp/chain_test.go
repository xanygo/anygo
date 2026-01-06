//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-06

package xcmp_test

import (
	"slices"
	"testing"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/xt"
)

func TestChain(t *testing.T) {
	var users1 = []user{user1, user2, user3, user4, user5}
	slices.SortFunc(users1, xcmp.Chain(
		// name 包含 han 的排在前面
		xcmp.StrContains[user](func(u user) string { return u.Name }, "han", true),

		// 大的排在前面
		xcmp.Compare[user, int](func(t user) int { return t.Age }, false),
	))
	want := []user{user2, user1, user4, user3, user5}
	xt.Equal(t, want, users1)
}

var user1 = user{
	Name:  "lilei",
	Age:   18,
	Grade: 3,
	Class: 1,
}
var user2 = user{
	Name:  "hanMeiMei",
	Age:   16,
	Grade: 3,
	Class: 1,
}
var user3 = user{
	Name:  "jay",
	Age:   12,
	Grade: 3,
	Class: 1,
}

var user4 = user{
	Name:  "tom",
	Age:   18,
	Grade: 3,
	Class: 1,
}
var user5 = user{
	Name:  "lee",
	Age:   12,
	Grade: 6,
	Class: 8,
}

type user struct {
	Name  string
	Age   int
	Grade int
	Class int
}
