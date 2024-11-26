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
		before, after := xstr.CutIndex(s, index, 1)
		fmt.Printf("before=%q after=%q\n", before, after)
	}
	printCut("abc", 1)  // before="a" after="c"
	printCut("abc", -1) // before="" after="abc"
	printCut("abc", 10) // before="abc" after=""

	// Output:
	// before="a" after="c"
	// before="" after="abc"
	// before="abc" after=""
}

func ExampleCutIndexBefore() {
	fmt.Println(xstr.CutIndexBefore("abcd", 2))  // ab
	fmt.Println(xstr.CutIndexBefore("abcd", 0))  //
	fmt.Println(xstr.CutIndexBefore("abcd", 1))  // a
	fmt.Println(xstr.CutIndexBefore("abcd", -1)) //
	fmt.Println(xstr.CutIndexBefore("abcd", 4))  // abcd

	// Output:
	// ab
	//
	// a
	//
	// abcd
}

func ExampleCutIndexAfter() {
	fmt.Println(xstr.CutIndexAfter("abcd", 2, 1))  // d
	fmt.Println(xstr.CutIndexAfter("abcd", 0, 1))  // bcd
	fmt.Println(xstr.CutIndexAfter("abcd", 1, 1))  // cd
	fmt.Println(xstr.CutIndexAfter("abcd", 4, 1))  //
	fmt.Println(xstr.CutIndexAfter("abcd", -1, 1)) // abcd

	// Output:
	// d
	// bcd
	// cd
	//
	// abcd
}

func ExampleCutLastByteN() {
	printCut := func(s string, c byte, n int) {
		before, after := xstr.CutLastByteN(s, c, n)
		fmt.Printf("before=%q after=%q\n", before, after)
	}

	printCut("/home/work/go/src/", '/', 2)  // before="/home/work/go" after="src/"
	printCut("/home/work/go/src/", '/', 10) // before="" after="/home/work/go/src/"

	// Output:
	// before="/home/work/go" after="src/"
	// before="" after="/home/work/go/src/"
}

func ExampleCutLastByteNBefore() {
	fmt.Println(xstr.CutLastByteNBefore("/home/work/go/src/", '/', 2))  // /home/work/go
	fmt.Println(xstr.CutLastByteNBefore("/home/work/go/src/", '/', 10)) //

	// Output:
	// /home/work/go
	//
}

func ExampleCutLastByteNAfter() {
	fmt.Println(xstr.CutLastByteNAfter("/home/work/go/src/", '/', 2))  // src/
	fmt.Println(xstr.CutLastByteNAfter("/home/work/go/src/", '/', 10)) // /home/work/go/src/

	// Output:
	// src/
	// /home/work/go/src/
}

func ExampleCutLastN() {
	printCut := func(s string, sub string, n int) {
		before, after := xstr.CutLastN(s, sub, n)
		fmt.Printf("before=%q after=%q\n", before, after)
	}
	printCut("abc-ab-ab-c", "ab", 1) // before="" after="c-ab-ab-c"
	printCut("abc-ab-ab-c", "ab", 2) // before="abc-" after="-ab-c"

	// Output:
	// before="" after="c-ab-ab-c"
	// before="abc-" after="-ab-c"
}

func ExampleToInts() {
	fmt.Println(xstr.ToInt32s("1,2,3,", ",")) // [1 2 3] <nil>
	fmt.Println(xstr.ToInt32s(",,3,4", ","))  // [3 4] <nil>

	// Output:
	// [1 2 3] <nil>
	// [3 4] <nil>
}

func ExampleToBools() {
	fmt.Println(xstr.ToBools("1,true,false,,", ","))

	// Output:
	// [true true false] <nil>
}

func ExampleSubstr() {
	fmt.Println(xstr.Substr("hello", 0, 1)) // h
	fmt.Println(xstr.Substr("hello", 0, 2)) // he
	fmt.Println(xstr.Substr("hello", 0, 5)) // hello

	// out off length
	fmt.Println(xstr.Substr("hello", 0, 6)) // hello

	// zero length
	fmt.Println(xstr.Substr("hello", 0, 0)) //

	// negative index
	fmt.Println(xstr.Substr("hello", -1, 1)) // o
	fmt.Println(xstr.Substr("hello", -1, 2)) // o
	fmt.Println(xstr.Substr("hello", -2, 2)) // lo
	fmt.Println(xstr.Substr("hello", -3, 3)) // llo

	// out of negative index
	fmt.Println(xstr.Substr("hello", -10, 3)) // hel

	// Output:
	// h
	// he
	// hello
	// hello
	//
	// o
	// o
	// lo
	// llo
	// hel
}

func ExampleToStrings() {
	fmt.Println(xstr.ToStrings(",1,2,3,  ,", ","))

	// Output:
	// [1 2 3]
}

func ExampleHasAnyPrefix() {
	fmt.Println(xstr.HasAnyPrefix("hello", "abc"))          // false
	fmt.Println(xstr.HasAnyPrefix("hello", "hello"))        // true
	fmt.Println(xstr.HasAnyPrefix("hello", "abc", "hello")) // true

	// Output:
	// false
	// true
	// true
}

func ExampleHasAnySuffix() {
	fmt.Println(xstr.HasAnySuffix("hello.js", ".a"))        // false
	fmt.Println(xstr.HasAnySuffix("hello.js", ".a", ".js")) // true
	fmt.Println(xstr.HasAnySuffix("hello.js", ".js"))       // true

	// Output:
	// false
	// true
	// true
}

func ExampleBytePairIndex() {
	str1 := `(hello(a,b,c,d(e,f),g),h) word(a,b)`

	leftIndex, rightIndex, ok := xstr.BytePairIndex(str1, '(', ')')
	fmt.Println("leftIndex=", leftIndex, "rightIndex=", rightIndex, "ok=", ok)

	substr := str1[leftIndex : rightIndex+1]
	fmt.Println("substr=", substr)

	// Output:
	// leftIndex= 0 rightIndex= 24 ok= true
	// substr= (hello(a,b,c,d(e,f),g),h)
}
