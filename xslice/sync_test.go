//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-24

package xslice

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestSync(t *testing.T) {
	ss := &Sync[int]{}
	ss.Grow(2)
	_, ok1 := ss.Head()
	xt.False(t, ok1)

	_, ok2 := ss.Tail()
	xt.False(t, ok2)

	ss.Append(1)
	v3, ok3 := ss.Head()
	xt.True(t, ok3)
	xt.Equal(t, v3, 1)

	v4, ok4 := ss.Tail()
	xt.True(t, ok4)
	xt.Equal(t, v4, 1)

	xt.Equal(t, ss.Load(), []int{1})
	ss.Clear()
	xt.Empty(t, ss.Load())

	ss.Store([]int{2})
	xt.Equal(t, ss.Load(), []int{2})

	ss.Grow(10)
	old := ss.Swap([]int{3})
	xt.Equal(t, old, []int{2})
	xt.Equal(t, ss.Load(), []int{3})

	old = append(old, 3)
	_ = old
	xt.Equal(t, ss.Load(), []int{3})

	ss.Delete(0, 1)
	xt.Empty(t, ss.Load())

	ss.Append(1, 10)
	ss.DeleteFunc(func(i int) bool {
		return i > 5
	})
	xt.Equal(t, ss.Load(), []int{1})
	ss.Insert(0, 2)
	xt.Equal(t, ss.Load(), []int{2, 1})

	ss1 := ss.Clone()
	xt.Equal(t, ss1.Load(), []int{2, 1})
}
