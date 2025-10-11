//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-28

package ztypes

import (
	"fmt"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xerror"
)

type Registry[K comparable, V any] interface {
	Register(name K, s V) error      // 注册，若重名会返回错误
	TryRegister(name K, s any) error // 注册，若重名或者类型不对会返回错误
	MustRegister(name K, s any)      // 注册，若重名或者类型不对会panic
	Deregister(name K) V             // 注销，若不存在则返回 nil

	Upsert(name K, s V) (old V) // 注册，若重名则替换并返回旧的

	Find(name K) (V, bool)
	Range(fn func(name K, s V) bool)
}

func NewRegistry[K comparable, V any]() Registry[K, V] {
	return &registry[K, V]{}
}

type registry[K comparable, V any] struct {
	db xmap.Sync[K, V]
}

func (si *registry[K, V]) TryRegister(name K, s any) error {
	if vv, ok := s.(V); ok {
		return si.Register(name, vv)
	}
	return fmt.Errorf("cannot register %v: %T", name, s)
}

func (si *registry[K, V]) MustRegister(name K, s any) {
	err := si.TryRegister(name, s)
	if err == nil {
		return
	}
	panic(err)
}

func (si *registry[K, V]) Register(name K, s V) error {
	_, loaded := si.db.LoadOrStore(name, s)
	if loaded {
		return fmt.Errorf("%w: %v", xerror.DuplicateKey, name)
	}
	return nil
}

func (si *registry[K, V]) Deregister(name K) V {
	old, _ := si.db.LoadAndDelete(name)
	return old
}

func (si *registry[K, V]) Upsert(name K, s V) (old V) {
	oldItem, loaded := si.db.LoadOrStore(name, s)
	if loaded {
		si.db.Store(name, s)
		return oldItem
	}
	return old
}

func (si *registry[K, V]) Find(name K) (V, bool) {
	return si.db.Load(name)
}

func (si *registry[K, V]) Range(fn func(name K, s V) bool) {
	si.db.Range(func(name K, value V) bool {
		return fn(name, value)
	})
}
