//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-02

package dsession

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xoption"
)

var protocols = &xmap.Sync[string, Starter]{}

func RegisterProtocol(protocol string, h Starter) error {
	if protocol == "" {
		return errors.New("protocol name is empty")
	}
	_, loaded := protocols.LoadOrStore(strings.ToUpper(protocol), h)
	if loaded {
		return fmt.Errorf("protocol %s already registered", protocol)
	}
	return nil
}

// FindProtocol 按照协议查找，若找不到会返回nil
func FindProtocol(protocol string) Starter {
	handler, _ := protocols.Load(strings.ToUpper(protocol))
	return handler
}

var registry = &xmap.Sync[string, FactoryFunc]{}

// RegisterFactory 注册自定义会话初始化工厂方法，
// 注册后可以在 service 配置的 SessionInit 段落使用
func RegisterFactory(name string, factory FactoryFunc) error {
	_, loaded := registry.LoadOrStore(name, factory)
	if loaded {
		return fmt.Errorf("factory %q already registered", name)
	}
	return nil
}

type FactoryFunc func(cfg map[string]any) (Starter, error)

func create(cfg *xoption.SessionStarterConfig) (Starter, error) {
	if cfg == nil {
		return nil, nil
	}
	fn, ok := registry.Load(cfg.Name)
	if !ok {
		return nil, fmt.Errorf("not found %q", cfg.Name)
	}
	return fn(cfg.Params)
}
