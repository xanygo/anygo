//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xsync

import (
	"errors"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestOnceDoErr(t *testing.T) {
	var once OnceDoErr
	var num int
	for i := 0; i < 3; i++ {
		got := once.Do(func() error {
			num++
			return errors.New("hello")
		})
		xt.Error(t, got)
		xt.Equal(t, "hello", got.Error())
		xt.Equal(t, 1, num)
	}
}

func TestOnceValueErr(t *testing.T) {
	var once OnceDoValueErr[string]
	var num int
	for i := 0; i < 3; i++ {
		got1, got2 := once.Do(func() (string, error) {
			num++
			return "ok", errors.New("hello")
		})
		xt.Error(t, got2)
		xt.Equal(t, "hello", got2.Error())
		xt.Equal(t, "ok", got1)
		xt.Equal(t, 1, num)
	}
}

func TestOnceValue(t *testing.T) {
	var num int
	one := OnceValue[int](func() int {
		num++
		return num
	})
	for i := 0; i < 3; i++ {
		xt.Equal(t, 1, one())
	}
}

func TestOnceValue2(t *testing.T) {
	var num1 int
	var num2 int
	once := OnceValue2[int, int](func() (int, int) {
		num1++
		num2 = num2 + 2
		return num1, num2
	})
	for i := 0; i < 3; i++ {
		v1, v2 := once()
		xt.Equal(t, 1, v1)
		xt.Equal(t, 2, v2)
	}
}

func TestOnceValue3(t *testing.T) {
	var num1 int
	var num2 int
	var num3 int
	once := OnceValue3[int, int, int](func() (int, int, int) {
		num1++
		num2 += 2
		num3 += 5
		return num1, num2, num3
	})
	for i := 0; i < 3; i++ {
		v1, v2, v3 := once()
		xt.Equal(t, 1, v1)
		xt.Equal(t, 2, v2)
		xt.Equal(t, 5, v3)
	}
}

func TestOnceValue4(t *testing.T) {
	var num1 int
	var num2 int
	var num3 int
	var num4 int
	once := OnceValue4[int, int, int, int](func() (int, int, int, int) {
		num1++
		num2 += 2
		num3 += 5
		num4 += 7
		return num1, num2, num3, num4
	})
	for i := 0; i < 3; i++ {
		v1, v2, v3, v4 := once()
		xt.Equal(t, 1, v1)
		xt.Equal(t, 2, v2)
		xt.Equal(t, 5, v3)
		xt.Equal(t, 7, v4)
	}
}
