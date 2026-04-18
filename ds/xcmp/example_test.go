//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-06

package xcmp_test

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xanygo/anygo/ds/xcmp"
)

type User struct {
	Name string
	Age  int
}

func ExampleChain() {
	users := []User{{Name: "John", Age: 18}, {Name: "tom-1", Age: 2}, {Name: "tom-2", Age: 4}}

	// 下列 SortFunc 实现了：
	// ORDER BY
	//  CASE WHEN Name LIKE '%tom%' THEN 0 ELSE 1 END ASC,
	//  Age DESC
	slices.SortFunc(users, xcmp.Chain[User](
		// Name 中包含 "tom" 的排在前面
		xcmp.TrueFront(func(t User) bool { return strings.Contains(t.Name, "tom") }),
		// 再按照 Age 大小降序
		xcmp.OrderDesc(func(t User) int { return t.Age }),
	))
	fmt.Println(users)
	// Output:
	// [{tom-2 4} {tom-1 2} {John 18}]
}

func ExampleOrderAsc() {
	var users []User // 待排序的数据

	// users = loadUsers()

	slices.SortFunc(users, xcmp.Chain[User](
		// 再按照 Age 大小升序
		xcmp.OrderAsc(func(t User) int { return t.Age }),
	))
}
