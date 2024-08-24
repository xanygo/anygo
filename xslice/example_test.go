//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice_test

import (
	"fmt"
	"github.com/xanygo/anygo/xslice"
	"strconv"
)

func ExampleMerge() {
	fmt.Println(xslice.Merge([]int{1}, []int{2, 3})) // [1 2 3]

	// Output:
	// [1 2 3]
}

func ExampleUnique() {
	fmt.Println(xslice.Unique([]int{1, 2, 1, 3})) // [1 2 3]

	// Output:
	// [1 2 3]
}

func ExampleContainsAny() {
	fmt.Println(xslice.ContainsAny([]int{1, 2, 3}, 1))    // true
	fmt.Println(xslice.ContainsAny([]int{1, 2, 3}, 4))    // false
	fmt.Println(xslice.ContainsAny([]int{1, 2, 3}, 3, 4)) // true

	// Output:
	// true
	// false
	// true
}

func ExampleToMap() {
	fmt.Println(xslice.ToMap([]int{1, 2, 3}, true)) // map[1:true 2:true 3:true]
	fmt.Println(xslice.ToMap([]int{1, 2, 3}, "ok")) // map[1:ok 2:ok 3:ok]

	// Output:
	// map[1:true 2:true 3:true]
	// map[1:ok 2:ok 3:ok]
}

func ExampleToAnys() {
	fmt.Printf("%#v\n", xslice.ToAnys([]int{1, 2, 3})) // []interface {}{1, 2, 3}

	// Output:
	// []interface {}{1, 2, 3}
}

func ExampleDeleteValue() {
	fmt.Println(xslice.DeleteValue([]int{1, 2, 3, 4}, 2, 4)) // [1 3]

	// Output:
	// [1 3]
}

func ExampleJoinFunc() {
	fmt.Println(xslice.JoinFunc([]int{1, 2}, strconv.Itoa, "-")) // 1-2

	fmt.Println(xslice.JoinFunc([]int{1, 2}, func(val int) string {
		return fmt.Sprintf("%02d", val)
	}, "-")) //  01-02

	// Output:
	// 1-2
	// 01-02
}

func ExampleJoin() {
	fmt.Println(xslice.Join([]int{1, 2}, "-")) // 1-2

	// Output:
	// 1-2
}
