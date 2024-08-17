//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely_test

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/safely"
)

func ExampleRunVoid() {
	safely.RunVoid(func() {
		panic("hello")
	}) // auto recovered

	fmt.Println("i'm ok")
	// Output:
	// i'm ok
}

func ExampleRun() {
	err := safely.Run(func() {
		panic("hello")
	}) // auto recovered

	fmt.Println("i'm ok")
	fmt.Println("err:", err == nil)
	// Output:
	// i'm ok
	// err: false
}

func ExampleRunCtxVoid() {
	safely.RunCtxVoid(context.Background(), func(ctx context.Context) {
		panic("hello")
	}) // auto recovered

	fmt.Println("i'm ok")
	// Output:
	// i'm ok
}

func ExampleRunCtx() {
	err := safely.Run(func() {
		panic("hello")
	}) // auto recovered

	fmt.Println("i'm ok")
	fmt.Println("err:", err == nil)
	// Output:
	// i'm ok
	// err: false
}
