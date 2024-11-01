//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-02

package xctx_test

import (
	"context"
	"fmt"

	"github.com/xanygo/anygo/xctx"
)

type ctxKey uint8

const ctxKey1 ctxKey = iota

func ExampleValues() {
	ctx1 := xctx.WithValues(context.Background(), ctxKey1, 1, 2)
	ctx2 := xctx.WithValues(ctx1, ctxKey1, 3, 4)

	// recursion = true
	// read from ctx2 and ctx1
	fmt.Println(xctx.Values[ctxKey, int](ctx2, ctxKey1, true)) // [1 2 3 4]

	// recursion = false
	// read from ctx2
	fmt.Println(xctx.Values[ctxKey, int](ctx2, ctxKey1, false)) // [3 4]

	// Output:
	// [1 2 3 4]
	// [3 4]
}
