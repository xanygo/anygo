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

func ExampleGet() {
	var m1 map[string]int

	// Get from nil map
	got1, ok1 := xmap.Get(m1, "k1")
	fmt.Println("k1=", got1, ok1) //  k1= 0 false

	m1 = map[string]int{"k1": 1}
	got2, ok2 := xmap.Get(m1, "k1")
	fmt.Println("k1=", got2, ok2) //  k1= 1 true

	// Output:
	// k1= 0 false
	// k1= 1 true
}

func ExampleGetDf() {
	var m1 map[string]int

	// Get from nil map
	fmt.Println("k1=", xmap.GetDf(m1, "k1", 0)) //  k1= 0
	fmt.Println("k1=", xmap.GetDf(m1, "k1", 1)) //  k1= 1

	m1 = map[string]int{"k1": 1}
	fmt.Println("k1=", xmap.GetDf(m1, "k1", 0)) //  k1= 1
	fmt.Println("k2=", xmap.GetDf(m1, "k2", 0)) //  k2= 0
	fmt.Println("k2=", xmap.GetDf(m1, "k2", 1)) //  k2= 1

	// Output:
	// k1= 0
	// k1= 1
	// k1= 1
	// k2= 0
	// k2= 1
}
