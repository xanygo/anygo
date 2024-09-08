//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-08

package xmap

type Tags[K comparable, V any, T comparable] struct {
	values map[K]*tagValue[V, T]
	index  map[T][]K
}

func (t *Tags[K, V, T]) Add(k K, v V, tags ...T) {
	if t.values == nil {
		t.values = make(map[K]*tagValue[V, T])
		t.index = make(map[T][]K)
	}
	t.values[k] = &tagValue[V, T]{
		Value: v,
		Tags:  tags,
	}
	for _, tag := range tags {
		t.index[tag] = append(t.index[tag], k)
	}
}

func (t *Tags[K, V, T]) Any(dk K, tags ...T) (v V) {
	if len(t.values) == 0 {
		return v
	}
	for _, tag := range tags {
		vs, ok := t.index[tag]
		if !ok || len(vs) == 0 {
			continue
		}
		return t.values[vs[0]].Value
	}
	val, ok := t.values[dk]
	if ok {
		return val.Value
	}
	for _, item := range t.values {
		return item.Value
	}
	return v
}

type tagValue[V any, T comparable] struct {
	Value V
	Tags  []T
}
