//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-17

package safely_test

import (
	"context"
	"fmt"
	"sync/atomic"

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

func ExampleOnRecovered() {
	var panicCount atomic.Int64
	safely.OnRecovered(func(ctx context.Context, re *safely.PanicErr) {
		panicCount.Add(1)
	})
}

func ExampleWrap() {
	fn1 := func() error {
		panic("hello")
	}
	fn2 := safely.Wrap(fn1)
	fmt.Println("err is nil:", fn2() == nil)
	// Output:
	// err is nil: false
}

func ExampleWrapVoid() {
	fn1 := func() error {
		panic("hello")
	}
	fn2 := safely.WrapVoid(fn1)
	fn2()
	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleWrapCtxVoid() {
	fn1 := func(ctx context.Context) {
		panic("hello")
	}
	fn2 := safely.WrapCtxVoid(fn1)
	fn2(context.Background())
	fmt.Println("ok")
	// Output:
	// ok
}
