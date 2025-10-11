//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-08

package xmap

// Tags 以 map  格式存储 < key: value & tags > 数据，并且支持使用 tags 查找数据。
// 非并发安全的
type Tags[K comparable, V any, T comparable] struct {
	values map[K]*tagValue[V, T]
	index  map[T][]K
}

func (t *Tags[K, V, T]) Set(key K, value V, tags ...T) {
	if t.values == nil {
		t.values = make(map[K]*tagValue[V, T])
		t.index = make(map[T][]K)
	}
	t.values[key] = &tagValue[V, T]{
		Value: value,
		Tags:  tags,
	}
	for _, tag := range tags {
		t.index[tag] = append(t.index[tag], key)
	}
}

// Any 先依次使用 tags 查找，如无则尝试查找 defaultKey，若也不存在，则返回任意一个
func (t *Tags[K, V, T]) Any(defaultKey K, tags ...T) (value V) {
	if len(t.values) == 0 {
		return value
	}
	for _, tag := range tags {
		vs, ok := t.index[tag]
		if !ok || len(vs) == 0 {
			continue
		}
		return t.values[vs[0]].Value
	}
	val, ok := t.values[defaultKey]
	if ok {
		return val.Value
	}
	for _, item := range t.values {
		return item.Value
	}
	return value
}

type tagValue[V any, T comparable] struct {
	Value V
	Tags  []T
}
