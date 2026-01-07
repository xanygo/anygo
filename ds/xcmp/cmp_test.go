//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

//go:build go1.26

package xcmp_test

import (
	"slices"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/xt"
)

func TestOrderCustomAsc(t *testing.T) {
	u1 := user{
		Name:  "han",
		Ctime: time.Now(),
	}
	u2 := user{
		Name:  "li",
		Ctime: time.Now().Add(time.Hour),
	}
	var users = []user{u1, u2}
	users1 := slices.Clone(users)
	slices.SortFunc(users1,
		xcmp.Chain(
			xcmp.OrderCustomDesc(func(t user) time.Time { return t.Ctime }),
		),
	)
	want := []user{u2, u1}
	xt.Equal(t, want, users1)

	users2 := slices.Clone(users)
	slices.SortFunc(users2,
		xcmp.ReverseChain(
			xcmp.OrderCustomAsc(func(t user) time.Time { return t.Ctime }),
		),
	)
	xt.Equal(t, want, users2)
}
