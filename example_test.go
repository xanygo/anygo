//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package anygo_test

import (
	"fmt"

	"github.com/xanygo/anygo"
	"io"
)

func ExampleTernary() {
	fmt.Println(anygo.Ternary(true, "v1", "v2")) // v1

	fmt.Println(anygo.Ternary(false, "v1", "v2")) // v2

	// Output:
	// v1
	// v2
}

func ExampleMust() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic:", err)
		}
	}()
	run := func() (int, error) {
		return 1, io.EOF
	}
	anygo.Must(run())

	// Output:
	// panic: EOF
}

func ExampleDo() {
	var called int
	err := anygo.DoThen(func() error {
		called += 1
		// do something
		return nil
	}).Then(func() error {
		called += 3
		// do something
		return io.EOF
	}).Then(func() error {
		// 由于前面 err != nil，所以此处的方法不会执行
		called += 5
		return nil
	}).Err()

	fmt.Println("called=", called)
	fmt.Println("err=", err)

	// Output:
	// called= 4
	// err= EOF
}
