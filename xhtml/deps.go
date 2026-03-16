//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-14

package xhtml

import (
	"sync"
)

// Deps 可用于 模版 和 Handler 之间传递是否需要资源 / 管理（js、css 等）依赖
type Deps struct {
	value sync.Map
}

// Use 标记需要使用
func (r *Deps) Use(names ...string) string {
	for _, name := range names {
		r.value.Store(name, true)
	}
	return ""
}

// Needs 返回是否需要
func (r *Deps) Needs(name string) bool {
	_, ok := r.value.Load(name)
	return ok
}

func (r *Deps) All() []string {
	var list []string
	r.value.Range(func(key, _ any) bool {
		list = append(list, key.(string))
		return true
	})
	return list
}
