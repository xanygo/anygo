//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xsync_test

import (
	"fmt"

	"github.com/xanygo/anygo/ds/xsync"
)

func ExampleOnceDoValue_Do() {
	var num int
	fn := func() int {
		num++
		return num
	}
	once := &xsync.OnceDoValue[int]{}
	fmt.Println("Done:", once.Done())
	for i := 0; i < 3; i++ {
		got := once.Do(fn)
		fmt.Println(got)
	}
	fmt.Println("Done:", once.Done())
	// Output:
	// Done: false
	// 1
	// 1
	// 1
	// Done: true
}

func ExampleOnceDoValue2_Done() {
	once := &xsync.OnceDoValue[int]{}
	fmt.Println("Done:", once.Done())
	got := once.Do(func() int {
		return 1
	})
	fmt.Println("Do:", got)
	fmt.Println("Done:", once.Done())

	// Output:
	// Done: false
	// Do: 1
	// Done: true
}

func ExampleOnceDoValue2_DoneValue() {
	once := &xsync.OnceDoValue[int]{}
	fmt.Print("DoneValue: ")
	fmt.Println(once.DoneValue())
	got := once.Do(func() int {
		return 1
	})
	fmt.Println("Do:", got)
	fmt.Print("DoneValue: ")
	fmt.Println(once.DoneValue())

	// Output:
	// DoneValue: false 0
	// Do: 1
	// DoneValue: true 1
}

func ExampleOnceDoValue2_Do() {
	var num1 int
	var num2 int
	fn := func() (int, int) {
		num1++
		num2 += 3
		return num1, num2
	}
	once := &xsync.OnceDoValue2[int, int]{}
	for i := 0; i < 3; i++ {
		v1, v2 := once.Do(fn)
		fmt.Println(v1, v2)
	}
	// Output:
	// 1 3
	// 1 3
	// 1 3
}

func ExampleOnceDoValue3_Do() {
	var num1 int
	var num2 int
	var num3 int
	fn := func() (int, int, int) {
		num1++
		num2 += 3
		num3 += 5
		return num1, num2, num3
	}
	once := &xsync.OnceDoValue3[int, int, int]{}
	for i := 0; i < 3; i++ {
		v1, v2, v3 := once.Do(fn)
		fmt.Println(v1, v2, v3)
	}
	// Output:
	// 1 3 5
	// 1 3 5
	// 1 3 5
}

func ExampleOnceDoValue4_Do() {
	var num1 int
	var num2 int
	var num3 int
	var num4 int
	fn := func() (int, int, int, int) {
		num1++
		num2 += 3
		num3 += 5
		num4 += 7
		return num1, num2, num3, num4
	}
	once := &xsync.OnceDoValue4[int, int, int, int]{}
	for i := 0; i < 3; i++ {
		v1, v2, v3, v4 := once.Do(fn)
		fmt.Println(v1, v2, v3, v4)
	}
	// Output:
	// 1 3 5 7
	// 1 3 5 7
	// 1 3 5 7
}

func ExampleOnceValue() {
	var num1 int
	fn := func() int {
		num1++
		return num1
	}

	once := xsync.OnceValue[int](fn)
	for i := 0; i < 3; i++ {
		fmt.Println(once())
	}
	// Output:
	// 1
	// 1
	// 1
}

func ExampleOnceValue2() {
	var num1 int
	var num2 int
	fn := func() (int, int) {
		num1++
		num2 += 3
		return num1, num2
	}

	once := xsync.OnceValue2[int, int](fn)
	for i := 0; i < 3; i++ {
		fmt.Println(once())
	}
	// Output:
	// 1 3
	// 1 3
	// 1 3
}

func ExampleGetBytesBuffer() {
	bf := xsync.GetBytesBuffer()
	bf.WriteString("hello")

	xsync.PutBytesBuffer(bf)
}
