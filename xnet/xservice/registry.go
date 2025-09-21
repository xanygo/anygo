//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"fmt"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xerror"
)

// Registry 下游服务管理器
type Registry interface {
	Register(s Service) error       // 注册，若重名会返回错误
	Deregister(name string) Service // 注销，若不存在则返回 nil

	Upsert(s Service) (old Service) // 注册，若重名则替换并返回旧的

	Find(name string) (Service, bool)
	Range(fn func(s Service) bool)
}

func NewRegistry() Registry {
	return &registryImpl{}
}

func DefaultRegistry() Registry {
	return defaultServiceRegistry
}

var defaultServiceRegistry = NewRegistry()

var _ Registry = (*registryImpl)(nil)

type registryImpl struct {
	db xmap.Sync[string, Service]
}

func (si *registryImpl) Register(s Service) error {
	_, loaded := si.db.LoadOrStore(s.Name(), s)
	if loaded {
		return fmt.Errorf("%w: %q", xerror.DuplicateKey, s.Name())
	}
	return nil
}

func (si *registryImpl) Deregister(name string) Service {
	old, _ := si.db.LoadAndDelete(name)
	return old
}

func (si *registryImpl) Upsert(s Service) (old Service) {
	old, loaded := si.db.LoadOrStore(s.Name(), s)
	if loaded {
		si.db.Store(s.Name(), s)
		return old
	}
	return nil
}

func (si *registryImpl) Find(name string) (Service, bool) {
	return si.db.Load(name)
}

func (si *registryImpl) Range(fn func(s Service) bool) {
	si.db.Range(func(_ string, value Service) bool {
		return fn(value)
	})
}
