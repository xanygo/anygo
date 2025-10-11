//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package anygo

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestTernary(t *testing.T) {
	fst.Equal(t, 1, Ternary(true, 1, 2))
	fst.Equal(t, 2, Ternary(false, 1, 2))
}

func TestMust(t *testing.T) {
	fst.Panic(t, func() {
		fn := func() (int, error) {
			panic("hello")
		}
		Must(fn())
	})
}

func TestMust1(t *testing.T) {
	fst.Panic(t, func() {
		Must1[int]((func() (int, error) {
			panic("hello")
		})())
	})

	fn1 := func() (int, error) {
		return 1, nil
	}
	fst.Equal(t, 1, Must1[int](fn1()))
}

func TestMust2(t *testing.T) {
	fst.Panic(t, func() {
		fn := func() (int, int, error) {
			panic("hello")
		}
		Must2(fn())
	})
	fn1 := func() (int, int, error) {
		return 1, 2, nil
	}
	v1, v2 := Must2[int, int](fn1())
	fst.Equal(t, 1, v1)
	fst.Equal(t, 2, v2)
}

func TestMust3(t *testing.T) {
	fst.Panic(t, func() {
		fn := func() (int, int, int, error) {
			panic("hello")
		}
		Must3(fn())
	})
	fn1 := func() (int, int, int, error) {
		return 1, 2, 3, nil
	}
	v1, v2, v3 := Must3[int, int, int](fn1())
	fst.Equal(t, 1, v1)
	fst.Equal(t, 2, v2)
	fst.Equal(t, 3, v3)
}
