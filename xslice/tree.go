package xslice

import (
	"cmp"
	"slices"
)

type TreeNode[T comparable, V any] struct {
	ID       T
	Value    V
	Parent   T
	Children []*TreeNode[T, V]
}

func ToTree[T comparable, V any](items []V, id func(V) T, parent func(V) T) []*TreeNode[T, V] {
	nodes := make(map[T]*TreeNode[T, V], len(items))
	for _, item := range items {
		node := &TreeNode[T, V]{
			ID:     id(item),
			Value:  item,
			Parent: parent(item),
		}

		nodes[node.ID] = node
	}

	var roots []*TreeNode[T, V]

	var zero T
	for _, node := range nodes {
		if node.Parent == zero {
			roots = append(roots, node)
			continue
		}

		if parentNode, ok := nodes[node.Parent]; ok {
			parentNode.Children = append(parentNode.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	return roots
}

func SortTreeNodes[T cmp.Ordered, V any](tree []*TreeNode[T, V], sort func(a, b T) int) {
	if Len(tree) == 0 {
		return
	}
	slices.SortStableFunc(tree, func(a, b *TreeNode[T, V]) int {
		return sort(a.ID, b.ID)
	})
	for _, item := range tree {
		SortTreeNodes(item.Children, sort)
	}
}
