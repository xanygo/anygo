package ztypes

type Named interface {
	Name() string
}

func Named2[K any, V any](name string) Named {
	return named2[K, V]{
		name: name,
	}
}

var _ Named = named2[string, string]{name: "a", a: "a", b: "b"}

type named2[A any, B any] struct {
	name string
	a    A
	b    B
}

func (n named2[A, B]) Name() string {
	return n.name
}
