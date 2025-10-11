//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-07

package tplfn

import "html/template"

type Iter[T any] interface {
	Next() T
}

type IterNextFunc[T any] func() T

func (f IterNextFunc[T]) Next() T {
	return f()
}

type ValueAttr interface {
	Value(v any) template.HTMLAttr
}

type ValueAttrFunc func(v any) template.HTMLAttr

func (ov ValueAttrFunc) Value(v any) template.HTMLAttr {
	return ov(v)
}
