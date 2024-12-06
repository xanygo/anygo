//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package tplfn

import (
	"fmt"
	"html/template"
	"math/rand/v2"
	"time"
)

func OptionSelected(value any) func(current any) template.HTMLAttr {
	valuesStr := fmt.Sprint(value)
	return func(current any) template.HTMLAttr {
		var selected string
		cstr := fmt.Sprint(current)
		if cstr == valuesStr {
			selected = " selected"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), selected)
		return template.HTMLAttr(code)
	}
}

func Checked(value any) func(current any) template.HTMLAttr {
	valuesStr := fmt.Sprint(value)
	return func(current any) template.HTMLAttr {
		var checked string
		cstr := fmt.Sprint(current)
		if cstr == valuesStr {
			checked = " checked"
		}
		code := fmt.Sprintf("value=%q%s", template.HTMLEscapeString(cstr), checked)
		return template.HTMLAttr(code)
	}
}

func EachOfIter(values ...any) Iter[any] {
	var index int
	total := len(values)
	next := func() any {
		val := values[index%total]
		index++
		return val
	}
	return IterNextFunc[any](next)
}

func DateTime(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format("2006-01-02 15:04:05")
}

func RandOfIter(values ...any) Iter[any] {
	next := func() (e any) {
		if len(values) == 0 {
			return e
		}
		return values[rand.IntN(len(values))]
	}
	return IterNextFunc[any](next)
}

type Iter[T any] interface {
	Next() T
}

type IterNextFunc[T any] func() T

func (f IterNextFunc[T]) Next() T {
	return f()
}
