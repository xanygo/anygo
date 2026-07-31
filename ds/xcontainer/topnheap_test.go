package xcontainer

import (
	"fmt"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestNewTopNHeap(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		xh := NewTopNHeap[int](2, func(head, add int) bool {
			return head < add // 保留最大值
		})
		xh.Add(1)
		xt.Equal(t, xh.Len(), 1)
		xt.Equal(t, xh.Items(), []int{1})
		xh.Add(2)
		xh.Add(3)

		xt.Equal(t, xh.Len(), 2)
		xt.Equal(t, xh.Sorted(), []int{3, 2})
	})

	t.Run("case 2", func(t *testing.T) {
		xh := NewTopNHeap[int](3, func(head, add int) bool {
			return head > add // 保留最小值
		})
		xh.Add(1)
		xt.Equal(t, xh.Len(), 1)
		xt.Equal(t, xh.Items(), []int{1})
		for i := 2; i < 10; i++ {
			xh.Add(i)
		}
		xt.Equal(t, xh.Len(), 3)
		xt.Equal(t, xh.Sorted(), []int{1, 2, 3})
	})
	t.Run("case 3", func(t *testing.T) {
		xh := NewTopNHeap[int](3, func(head, add int) bool {
			return head < add // 保留最大值
		})
		xh.Add(1)
		xt.Equal(t, xh.Len(), 1)
		xt.Equal(t, xh.Items(), []int{1})
		for i := 2; i < 10; i++ {
			xh.Add(i)
		}
		xt.Equal(t, xh.Len(), 3)
		xt.Equal(t, xh.Sorted(), []int{9, 8, 7})
	})
}

func ExampleNewTopNHeap() {
	xh := NewTopNHeap[int](2, func(a, b int) bool {
		return a < b
	})
	for i := 0; i < 5; i++ {
		xh.Add(i)
	}
	// Output:
	// values: [4 3]

	fmt.Println("values:", xh.Sorted())
}
