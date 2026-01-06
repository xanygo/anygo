//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-06

package xcmp_test

import (
	"fmt"
	"slices"

	"github.com/xanygo/anygo/ds/xcmp"
)

func ExampleChain() {
	type User struct {
		Name string
		Age  int
	}
	users := []User{
		{
			Name: "John",
			Age:  18,
		},
		{
			Name: "tom-1",
			Age:  2,
		},
		{
			Name: "tom-2",
			Age:  4,
		},
	}
	slices.SortFunc(users, xcmp.Chain[User](
		// Name 中包含 "tom" 的排在前面
		xcmp.StrContains(func(t User) string { return t.Name }, "tom", true),
		// 再按照 Age 大小降序
		xcmp.Compare(func(t User) int { return t.Age }, false),
	))
	fmt.Println(users)
	// Output:
	// [{tom-2 4} {tom-1 2} {John 18}]
}
