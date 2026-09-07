package xcontainer

import (
	"slices"
	"sort"
)

// NewTopNHeap 创建一个保留  topN 的最小堆
//
//	n:  堆的最大容量
//	less: 比较函数，若 head(堆顶的值) < add(新的元素) 则将 add 添加到堆中。
//	    若 less 返回的是  head < add ，Heap 保留最大的 n 个值
//	    若 less 返回的是  head > add ，Heap 保留最小的 n 个值
func NewTopNHeap[T any](n int, less func(head, add T) bool) *TopNHeap[T] {
	return &TopNHeap[T]{
		data: make([]T, 0, n),
		n:    n,
		less: less,
	}
}

// TopNHeap 保留排序最大的 N 个元素
// 内部使用小根堆
type TopNHeap[T any] struct {
	data []T
	n    int
	less func(head, add T) bool // head < add
}

// Add 添加元素
func (h *TopNHeap[T]) Add(v T) {
	if h.n <= 0 {
		return
	}

	if len(h.data) < h.n {
		h.data = append(h.data, v)
		h.up(len(h.data) - 1)
		return
	}

	// 当前元素比最小元素大，替换
	if h.less(h.data[0], v) {
		h.data[0] = v
		h.down(0)
	}
}

// Items 返回当前保存的数据(无序的)
func (h *TopNHeap[T]) Items() []T {
	return h.data
}

// Sorted 返回已排序的结果。
// 若 less 返回的是 head < add，结果是最大的N个值，倒序(Sort Desc)
// 若 less 返回的是 head > add，结果是最小的N个值，正序(Sort Asc)
func (h *TopNHeap[T]) Sorted() []T {
	result := slices.Clone(h.data)
	sort.Slice(result, func(i, j int) bool {
		return h.less(result[j], result[i])
	})

	return result
}

func (h *TopNHeap[T]) Len() int {
	return len(h.data)
}

func (h *TopNHeap[T]) up(i int) {
	for {
		p := (i - 1) / 2
		if i == 0 || !h.less(h.data[i], h.data[p]) {
			break
		}

		h.data[i], h.data[p] = h.data[p], h.data[i]
		i = p
	}
}

func (h *TopNHeap[T]) down(i int) {
	n := len(h.data)

	for {
		left := i*2 + 1
		if left >= n {
			break
		}

		small := left
		right := left + 1

		if right < n && h.less(h.data[right], h.data[left]) {
			small = right
		}

		if !h.less(h.data[small], h.data[i]) {
			break
		}

		h.data[i], h.data[small] = h.data[small], h.data[i]
		i = small
	}
}
