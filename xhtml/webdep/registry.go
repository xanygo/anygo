package webdep

import (
	"errors"
	"fmt"
	"slices"
	"sync"
)

type AssetGetter interface {
	Get(name string) (Asset, bool)
}

// AssetGetters 将多个 AssetGetter 合并为一个
//
// 如可以将全局的 AssetGetter 和 本地局部的合并在一起。实现本地局部的可以覆盖全局的效果。
type AssetGetters []AssetGetter

// Get 获取指定名称的资源。
//
// 逆向遍历 AssetGetters，读取数据。
func (as AssetGetters) Get(name string) (Asset, bool) {
	for i := len(as) - 1; i >= 0; i-- {
		if item, ok := as[i].Get(name); ok {
			return item, true
		}
	}
	return Asset{}, false
}

// DefaultRegistry 全局的，默认的 Registry
var DefaultRegistry = NewRegistry()

// NewRegistry 创建资源注册表。
func NewRegistry() *Registry {
	return &Registry{
		assets: make(map[string]Asset),
	}
}

var _ AssetGetter = (*Registry)(nil)

// Registry 管理 Web 资源注册信息。
type Registry struct {
	mu     sync.RWMutex
	assets map[string]Asset
	order  []string
}

// Register 注册一个或多个资源。
// 同名资源会覆盖原有定义，但保留首次注册顺序。所有资源验证通过后才会统一注册。
func (r *Registry) Register(assets ...Asset) error {
	for _, asset := range assets {
		if err := validateAsset(asset); err != nil {
			return err
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, asset := range assets {
		if _, ok := r.assets[asset.Name]; !ok {
			r.order = append(r.order, asset.Name)
		}

		r.assets[asset.Name] = asset.Clone()
	}

	return nil
}

// Get 获取指定名称的资源。
// 返回副本，避免外部修改内部数据。
func (r *Registry) Get(name string) (Asset, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	asset, ok := r.assets[name]
	if !ok {
		return Asset{}, false
	}

	return asset.Clone(), true
}

// Has 判断资源是否已注册。
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.assets[name]
	return ok
}

// Delete 删除指定资源。
// 删除后再次注册时，资源会排在注册顺序末尾。
func (r *Registry) Delete(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.assets[name]; !ok {
		return false
	}

	delete(r.assets, name)
	r.order = slices.DeleteFunc(r.order, func(item string) bool {
		return item == name
	})
	return true
}

// All 按注册顺序返回所有资源。
func (r *Registry) All() []Asset {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Asset, 0, len(r.order))

	for _, name := range r.order {
		if asset, ok := r.assets[name]; ok {
			list = append(list, asset.Clone())
		}
	}

	return list
}

// Names 按注册顺序返回所有资源名称。
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Clone(r.order)
}

// Len 返回已注册资源数量。
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.assets)
}

// validateAsset 验证资源定义。
func validateAsset(asset Asset) error {
	if asset.Name == "" {
		return errors.New("asset name cannot be empty")
	}

	if asset.URL == "" {
		return fmt.Errorf("asset %q: URL cannot be empty", asset.Name)
	}

	if asset.Type != AssetJS && asset.Type != AssetCSS {
		return fmt.Errorf("asset %q: unsupported type %q", asset.Name, asset.Type)
	}

	if asset.CrossOrigin != "" &&
		asset.CrossOrigin != CrossOriginAnonymous &&
		asset.CrossOrigin != CrossOriginUseCredentials {
		return fmt.Errorf("asset %q: invalid crossorigin %q", asset.Name, asset.CrossOrigin)
	}

	if asset.Type == AssetCSS && (asset.Defer || asset.Async) {
		return fmt.Errorf("asset %q: defer and async are only supported for JS", asset.Name)
	}

	if asset.Type == AssetJS && asset.Media != "" {
		return fmt.Errorf("asset %q: media is only supported for CSS", asset.Name)
	}

	for _, dep := range asset.DependsOn {
		if dep == "" {
			return fmt.Errorf("asset %q: dependency name cannot be empty", asset.Name)
		}
	}

	return nil
}
