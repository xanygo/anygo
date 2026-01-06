//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xmap_test

import (
	"fmt"
	"sort"

	"github.com/xanygo/anygo/ds/xmap"
)

func ExampleOrdered_Values() {
	mp := &xmap.OrderedSync[string, int]{}
	mp.Set("k0", 0)
	mp.Set("k1", 1)
	mp.Set("k2", 2)

	fmt.Println("values:", mp.Values())

	// Output:
	// values: [0 1 2]
}

func ExampleOrdered_Keys() {
	mp := &xmap.OrderedSync[string, int]{}
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

func ExampleFilter() {
	m := map[int]int{
		0: 0,
		1: 1,
		2: 2,
		3: 3,
	}
	result := xmap.Filter(m, func(k int, v int, ok int) bool {
		return v%2 == 0
	})
	fmt.Println(result)

	// Output:
	// map[0:0 2:2]
}

func ExampleFilterKeys() {
	m := map[int]int{
		0: 10,
		1: 11,
		2: 22,
		3: 33,
	}
	result := xmap.FilterKeys(m, func(k int, v int, ok int) bool {
		return v%2 == 0
	})
	sort.Ints(result)

	fmt.Println(result)

	// Output:
	// [0 2]
}

func ExampleFilterValues() {
	m := map[int]int{
		0: 10,
		1: 11,
		2: 22,
		3: 33,
	}
	result := xmap.FilterValues(m, func(k int, v int, ok int) bool {
		return v%2 == 0
	})
	sort.Ints(result)

	fmt.Println(result)

	// Output:
	// [10 22]
}

func ExampleKeysMiss() {
	mp := map[string]int{"a": 1, "b": 2}
	keys := []string{"a", "b", "c"}
	miss := xmap.KeysMiss(mp, keys)

	fmt.Println("miss=", miss)
	// Output:
	// miss= [c]
}

func ExampleRange() {
	mp := map[string]any{"a": 1, "b": int64(2)}

	var count1 int

	// 遍历出所有类型为 int 类型的 k-v 项
	matched := xmap.Range[string, int](mp, func(key string, val int) bool {
		count1 += val
		return true
	})
	fmt.Println("matched=", matched, ",count:", count1) // matched= 1 ,count: 1

	count1 = 0
	// value 使用 any，可以匹配所有类型
	matched = xmap.Range[string, any](mp, func(key string, val any) bool {
		switch rv := val.(type) {
		case int:
			count1 += rv
		case int64:
			count1 += int(rv)
		}
		return true
	})
	fmt.Println("matched=", matched, ",count:", count1) // matched= 2 ,count: 3

	// Output:
	// matched= 1 ,count: 1
	// matched= 2 ,count: 3
}
