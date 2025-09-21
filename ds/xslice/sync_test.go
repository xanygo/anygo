//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestSync(t *testing.T) {
	ss := &Sync[int]{}
	ss.Grow(2)
	_, ok1 := ss.Head()
	fst.False(t, ok1)

	_, ok2 := ss.Tail()
	fst.False(t, ok2)

	ss.Append(1)
	v3, ok3 := ss.Head()
	fst.True(t, ok3)
	fst.Equal(t, 1, v3)

	v4, ok4 := ss.Tail()
	fst.True(t, ok4)
	fst.Equal(t, 1, v4)

	fst.Equal(t, []int{1}, ss.Load())
	ss.Clear()
	fst.Empty(t, ss.Load())

	ss.Store([]int{2})
	fst.Equal(t, []int{2}, ss.Load())

	ss.Grow(10)
	old := ss.Swap([]int{3})
	fst.Equal(t, []int{2}, old)
	fst.Equal(t, []int{3}, ss.Load())

	old = append(old, 3)
	_ = old
	fst.Equal(t, []int{3}, ss.Load())

	ss.Delete(0, 1)
	fst.Empty(t, ss.Load())

	ss.Append(1, 10)
	ss.DeleteFunc(func(i int) bool {
		return i > 5
	})
	fst.Equal(t, []int{1}, ss.Load())
	ss.Insert(0, 2)
	fst.Equal(t, []int{2, 1}, ss.Load())

	ss1 := ss.Clone()
	fst.Equal(t, []int{2, 1}, ss1.Load())
}
