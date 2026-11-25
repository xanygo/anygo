//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-14

package webdep

import (
	"fmt"
	"html/template"
	"slices"
	"strings"
	"sync"
)

func NewDeps(registry AssetGetter) *Deps {
	return &Deps{
		used:     make(map[string]struct{}, 6),
		registry: registry,
	}
}

// Deps 可用于 模版 和 Handler 之间传递是否需要资源 / 管理（js、css 等）依赖
type Deps struct {
	registry AssetGetter

	mu        sync.RWMutex
	used      map[string]struct{}
	usedOrder []string
}

// Use 标记需要使用
//
// 不会检查 registry 中对应的资源是否存在
// 但是在调用 Render 方法的时候会检查资源必须存在
//
// 若 name 是以 "|" 开头，Render 时会忽略掉，调用可以使用 Needs 读取到
func (r *Deps) Use(names ...string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.used == nil {
		r.used = make(map[string]struct{}, max(len(names), 6))
	}

	for _, name := range names {
		if _, ok := r.used[name]; ok {
			continue
		}

		r.used[name] = struct{}{}
		r.usedOrder = append(r.usedOrder, name)
	}

	return ""
}

const tagPrefix = "Tag/"

func (r *Deps) UseTag(tags ...string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.used == nil {
		r.used = make(map[string]struct{}, max(len(tags), 6))
	}

	for _, tag := range tags {
		name := tagPrefix + tag
		if _, ok := r.used[name]; ok {
			continue
		}

		r.used[name] = struct{}{}
		r.usedOrder = append(r.usedOrder, name)
	}

	return ""
}

// Needs 使用资源名称( Name )判断资源是否被使用(不会检查 registry 中对应的资源是否存在)
func (r *Deps) Needs(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.used) == 0 {
		return false
	}

	_, ok := r.used[name]
	return ok
}

// NeedsTag 使用资源标签( Tag )判断资源是否被使用(不会检查 registry 中对应的资源是否存在)
func (r *Deps) NeedsTag(tag string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.used) == 0 {
		return false
	}
	name := tagPrefix + tag
	_, ok := r.used[name]
	return ok
}

func (r *Deps) All() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Clone(r.usedOrder)
}

// assets 获取当前页面需要输出的资源。
// 自动展开 DependsOn，并保证依赖资源优先输出。
// 重复依赖只输出一次。
func (r *Deps) assets() ([]Asset, error) {
	r.mu.RLock()
	names := slices.Clone(r.usedOrder)
	r.mu.RUnlock()

	if len(names) == 0 {
		return nil, nil
	}

	registry := r.registry
	if registry == nil {
		registry = DefaultRegistry
	}

	var result []Asset

	visited := make(map[string]struct{}, len(names))
	visiting := make(map[string]struct{}, 2)

	var visit func(string) error

	visit = func(name string) error {
		if _, ok := visited[name]; ok {
			return nil
		}

		if _, ok := visiting[name]; ok {
			return fmt.Errorf("circular asset dependency detected: %q", name)
		}

		if tag, ok := strings.CutPrefix(name, tagPrefix); ok {
			for _, a := range registry.GetByTag(tag) {
				if err := visit(a.Name); err != nil {
					return err
				}
			}
			return nil
		}

		asset, ok := registry.GetByName(name)
		if !ok {
			return fmt.Errorf("asset %q is not registered", name)
		}

		visiting[name] = struct{}{}

		for _, dep := range asset.Depends {
			if err := visit(dep); err != nil {
				return err
			}
		}

		delete(visiting, name)
		visited[name] = struct{}{}

		result = append(result, asset)

		return nil
	}

	for _, name := range names {
		// 特殊的名字，忽略
		if strings.HasPrefix(name, "|") {
			continue
		}
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Render 输出 HTML 资源标签，需要对应的资源在 AssetGetter 中存在，否则会报错
func (r *Deps) Render() (template.HTML, error) {
	assets, err := r.assets()
	if err != nil {
		return "", err
	}
	if len(assets) == 0 {
		return "", nil
	}

	b := &strings.Builder{}
	for _, asset := range assets {
		asset.writeBuilder(b)
		b.WriteByte('\n')
	}

	return template.HTML(b.String()), nil
}
