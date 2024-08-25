//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package anygo_test

import (
	"fmt"

	"github.com/xanygo/anygo"
)

func ExampleTernary() {
	fmt.Println(anygo.Ternary(true, "v1", "v2")) // v1

	fmt.Println(anygo.Ternary(false, "v1", "v2")) // v2

	// Output:
	// v1
	// v2
}
