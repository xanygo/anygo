//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-27

package xruntime_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/xanygo/anygo/internal/xruntime"
	"github.com/xanygo/anygo/xt"
)

func panic1() {
	var m map[string]any
	//lint:ignore SA5000 test panic
	m["a"] = 1
	log.Println(m)
}

func panic2() {
	a := []int{1, 2, 3}
	_ = a[3]
}

func panic3() {
	var a any = []int{1}
	var b any = []int{1}

	fmt.Println(a == b)
}

func TestPanicCaller(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		defer func() {
			re := recover()
			xt.NotEmpty(t, re)
			file, line, fn := xruntime.PanicCaller(1)
			xt.Contains(t, file, "aller_test.go")
			xt.Equal(t, line, 19)
			xt.Contains(t, fn, "panic1")
		}()
		panic1()
	})

	t.Run("case 2", func(t *testing.T) {
		defer func() {
			re := recover()
			xt.NotEmpty(t, re)
			file, line, fn := xruntime.PanicCaller(1)
			xt.Contains(t, file, "aller_test.go")
			xt.Equal(t, line, 25)
			xt.Contains(t, fn, "panic2")
		}()
		panic2()
	})

	t.Run("case 3", func(t *testing.T) {
		defer func() {
			re := recover()
			xt.NotEmpty(t, re)
			file, line, fn := xruntime.PanicCaller(1)
			xt.Contains(t, file, "aller_test.go")
			xt.Equal(t, line, 32)
			xt.Contains(t, fn, "panic3")
		}()
		panic3()
	})
}
