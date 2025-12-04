//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-04

package trustip

import (
	"net"
	"slices"
	"strings"
	"sync"
)

type Manager struct {
	mu   sync.RWMutex
	nets []*net.IPNet
}

// New 创建新的 Manager 实例
func New() *Manager {
	return &Manager{
		nets: make([]*net.IPNet, 0),
	}
}

// Set 替换所有可信代理
func (m *Manager) Set(cidrs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	newNets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		newNets = append(newNets, network)
	}
	m.nets = newNets
	return nil
}

func (m *Manager) MustSet(cidrs []string) {
	err := m.Set(cidrs)
	if err != nil {
		panic(err)
	}
}

// Add 追加一个可信代理,如 127.0.0.1/32
func (m *Manager) Add(cidrs ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := make(map[string]struct{}, len(m.nets))
	for _, cidr := range m.nets {
		index[cidr.String()] = struct{}{}
	}
	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		key := network.String()
		if _, ok := index[key]; ok {
			continue
		}
		m.nets = append(m.nets, network)
		index[key] = struct{}{}
	}
	return nil
}

func (m *Manager) MustAdd(cidrs ...string) {
	err := m.Add(cidrs...)
	if err != nil {
		panic(err)
	}
}

// Remove 删除一个 IPNet,如 127.0.0.1/32
func (m *Manager) Remove(cidr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	m.nets = slices.DeleteFunc(m.nets, func(ipNet *net.IPNet) bool {
		return ipNet.String() == network.String()
	})
	return nil
}

func (m *Manager) RemoveIPNet(ipNet *net.IPNet) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	var found bool
	m.nets = slices.DeleteFunc(m.nets, func(e *net.IPNet) bool {
		ret := ipNet.String() == e.String()
		if ret {
			found = true
		}
		return ret
	})
	return found
}

// List 返回所有可信 IPNet
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cidrs := make([]string, 0, len(m.nets))
	for _, n := range m.nets {
		cidrs = append(cidrs, n.String())
	}
	return cidrs
}

// IsTrusted 判断一个 IP 是否属于可信列表
func (m *Manager) IsTrusted(ip net.IP) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, n := range m.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
