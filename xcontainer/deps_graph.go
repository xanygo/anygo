package xcontainer

import (
	"fmt"
	"sync"
)

// DepGraph 并发安全的有向依赖图。
//
// 图中的边 a -> b 表示 a 依赖 b。
//
// 一个节点可以依赖多个节点，并且依赖关系可以有任意层级,添加依赖关系时，DepGraph 会检查是否会形成循环依赖。
type DepGraph[T comparable] struct {
	mu   sync.RWMutex
	deps map[T]map[T]struct{}
}

// NewDepGraph 创建一个空的依赖图。
func NewDepGraph[T comparable]() *DepGraph[T] {
	return &DepGraph[T]{
		deps: make(map[T]map[T]struct{}),
	}
}

// Add 添加依赖关系。
//
// Add(a, b, c) 表示：
//
//	a -> b
//	a -> c
//
//	如果添加任意一个依赖关系后会形成循环依赖，则不会添加本次调用中的任何依赖，并返回错误。
//	如果依赖关系已经存在，则忽略该依赖，不会返回错误。
//	被依赖的节点不需要事先通过 Add 添加。
func (g *DepGraph[T]) Add(a T, deps ...T) error {
	if len(deps) == 0 {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// 先检查所有依赖，确保本次操作具有原子性：如果任意一个依赖会形成环，则一个都不添加。
	for _, dep := range deps {
		if a == dep {
			return fmt.Errorf("cyclic dependency detected: %v -> %v", a, dep)
		}

		if g.reachableLocked(dep, a) {
			return fmt.Errorf("cyclic dependency detected: %v -> %v", a, dep)
		}
	}

	if g.deps == nil {
		g.deps = make(map[T]map[T]struct{})
	}

	set := g.deps[a]
	if set == nil {
		set = make(map[T]struct{}, len(deps))
		g.deps[a] = set
	}

	for _, dep := range deps {
		set[dep] = struct{}{}
	}

	return nil
}

// Remove 删除 a 对 dep 的依赖关系。 如果依赖关系不存在，则返回 false。
func (g *DepGraph[T]) Remove(a, dep T) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	set := g.deps[a]
	if set == nil {
		return false
	}

	if _, ok := set[dep]; !ok {
		return false
	}

	delete(set, dep)

	if len(set) == 0 {
		delete(g.deps, a)
	}

	return true
}

// RemoveNode 删除节点及其所有依赖关系，同时会删除其他节点对 a 的依赖关系。
func (g *DepGraph[T]) RemoveNode(a T) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.deps, a)

	for node, deps := range g.deps {
		delete(deps, a)

		if len(deps) == 0 {
			delete(g.deps, node)
		}
	}
}

// Has 判断 a 是否直接依赖 dep。
//
// 例如存在 A -> B -> C，则：
//
//	Has(A, B) == true
//	Has(A, C) == false
func (g *DepGraph[T]) Has(a, dep T) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	_, ok := g.deps[a][dep]
	return ok
}

// DependsOn 判断 a 是否直接或间接依赖 dep。
//
// 例如存在 A -> B -> C 则：
//
//	DependsOn(A, B) == true
//	DependsOn(A, C) == true
//	DependsOn(A, D) == false
func (g *DepGraph[T]) DependsOn(a, dep T) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.reachableLocked(a, dep)
}

// Deps 返回 a 的所有直接依赖。
//
// 返回的 map 是副本，调用者可以安全修改。
// 如果 a 没有任何依赖，则返回 nil。
func (g *DepGraph[T]) Deps(a T) map[T]struct{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	set := g.deps[a]
	if len(set) == 0 {
		return nil
	}

	result := make(map[T]struct{}, len(set))
	for dep := range set {
		result[dep] = struct{}{}
	}

	return result
}

// Len 返回图中存在依赖关系的节点数量。
//
// 注意：没有任何依赖关系的节点不会单独存储，因此不会计入 Len
func (g *DepGraph[T]) Len() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.deps)
}

// Clear 删除所有依赖关系。
func (g *DepGraph[T]) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.deps = make(map[T]map[T]struct{})
}

// reachableLocked 判断 from 是否可以沿依赖关系到达 target
func (g *DepGraph[T]) reachableLocked(from, target T) bool {
	if from == target {
		return true
	}

	visited := make(map[T]struct{})
	stack := []T{from}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if n == target {
			return true
		}

		if _, ok := visited[n]; ok {
			continue
		}
		visited[n] = struct{}{}

		for dep := range g.deps[n] {
			if _, ok := visited[dep]; !ok {
				stack = append(stack, dep)
			}
		}
	}

	return false
}
