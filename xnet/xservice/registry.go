//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xservice

import (
	"fmt"
	"strings"

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

func init() {
	RegisterRegistry("default", DefaultRegistry())
}

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

var registries = &xmap.Sync[string, Registry]{}

func RegisterRegistry(name string, r Registry) error {
	_, loaded := registries.LoadOrStore(name, r)
	if loaded {
		return fmt.Errorf("registry %q already exists", name)
	}
	return nil
}

func UnregisterRegistry(name string) bool {
	_, ok := registries.LoadAndDelete(name)
	return ok
}

func FindRegistry(name string) (Registry, error) {
	v, ok := registries.Load(name)
	if ok {
		return v, nil
	}
	return nil, fmt.Errorf("registry %q %w", name, xerror.NotFound)
}

// FindService 查找 service
//
// 若 name 不包含 / ，从 DefaultRegistry()
// 若 name 包含 /, 如 other/name, 则先使用 FindRegistry("other") 找到对应的 Registry，然后查找 name
// 若 name == “dummy”， 则直接返回 DummyService
func FindService(name string) (Service, error) {
	if name == Dummy {
		return DummyService(), nil
	}
	pre, after, found := strings.Cut(name, "/")
	if !found {
		ser, ok := DefaultRegistry().Find(name)
		if ok {
			return ser, nil
		}
		return nil, fmt.Errorf("service %q %w", name, xerror.NotFound)
	}
	reg, err := FindRegistry(pre)
	if err != nil {
		return nil, err
	}
	ser, ok := reg.Find(after)
	if ok {
		return ser, nil
	}
	return nil, fmt.Errorf("service %q %w in registry %q ", after, xerror.NotFound, pre)
}
