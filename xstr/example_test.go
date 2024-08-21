//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xstr_test

import (
	"fmt"
	"github.com/xanygo/anygo/xstr"
)

func ExampleIndexN() {
	fmt.Println(xstr.IndexN("abc/abc/abc", "abc", 1)) // 0
	fmt.Println(xstr.IndexN("abc/abc/abc", "abc", 2)) // 4
	fmt.Println(xstr.IndexN("abc/abc/abc", "abc", 3)) // 8
	fmt.Println(xstr.IndexN("abc/abc/abc", "abc", 4)) // -1

	// Output:
	// 0
	// 4
	// 8
	// -1
}

func ExampleLastIndexN() {
	fmt.Println(xstr.LastIndexN("abc/abc/abc", "abc", 1)) // 8
	fmt.Println(xstr.LastIndexN("abc/abc/abc", "abc", 2)) // 4
	fmt.Println(xstr.LastIndexN("abc/abc/abc", "abc", 3)) // 0
	fmt.Println(xstr.LastIndexN("abc/abc/abc", "abc", 4)) // -1

	// Output:
	// 8
	// 4
	// 0
	// -1
}

func ExampleIndexByteN() {
	fmt.Println(xstr.IndexByteN("abc/abc/abc", 'a', 1)) // 0
	fmt.Println(xstr.IndexByteN("abc/abc/abc", 'a', 2)) // 4
	fmt.Println(xstr.IndexByteN("abc/abc/abc", 'a', 3)) // 8
	fmt.Println(xstr.IndexByteN("abc/abc/abc", 'a', 4)) // -1

	// Output:
	// 0
	// 4
	// 8
	// -1
}

func ExampleLastIndexByteN() {
	fmt.Println(xstr.LastIndexByteN("abc/abc/abc", 'a', 1)) // 8
	fmt.Println(xstr.LastIndexByteN("abc/abc/abc", 'a', 2)) // 4
	fmt.Println(xstr.LastIndexByteN("abc/abc/abc", 'a', 3)) // 0
	fmt.Println(xstr.LastIndexByteN("abc/abc/abc", 'a', 4)) // -1

	// Output:
	// 8
	// 4
	// 0
	// -1
}

func ExampleCutIndex() {
	printCut := func(s string, index int) {
		before, after := xstr.CutIndex(s, index)
		fmt.Printf("before=%q after=%q\n", before, after)
	}
	printCut("hello", 1)  // before="h" after="ello"
	printCut("hello", -1) // before="" after="hello"
	printCut("hello", 10) // before="hello" after=""

	// Output:
	// before="h" after="ello"
	// before="" after="hello"
	// before="hello" after=""
}
