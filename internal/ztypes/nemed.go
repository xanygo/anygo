package ztypes

import "reflect"

type Named interface {
	Name() string
}

func Named2[K any, V any](name string) Named {
	return named2{
		name: name,
		a:    reflect.TypeFor[K](),
		b:    reflect.TypeFor[V](),
	}
}

type named2 struct {
	name string
	a    reflect.Type
	b    reflect.Type
}

func (n named2) Name() string {
	return n.name
}
