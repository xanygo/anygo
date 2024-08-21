//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xmap_test

import (
	"fmt"
	"github.com/xanygo/anygo/xmap"
)

func ExampleOrdered_Values() {
	mp := &xmap.Ordered[string, int]{}
	mp.Set("k0", 0)
	mp.Set("k1", 1)
	mp.Set("k2", 2)
	fmt.Println("values:", mp.Values())
	// Output:
	// values: [0 1 2]
}

func ExampleOrdered_Keys() {
	mp := &xmap.Ordered[string, int]{}
	mp.Set("k0", 0)
	mp.Set("k1", 1)
	mp.Set("k2", 2)
	fmt.Println("keys:", mp.Keys())
	// Output:
	// keys: [k0 k1 k2]
}
