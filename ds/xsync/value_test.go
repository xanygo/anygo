//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-18

package xsync

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/fsgo/fst"
)

func TestValueErr(t *testing.T) {
	checkErrValue := func(t *testing.T, ev *Value[error]) {
		fst.True(t, ev.CompareAndSwap(nil, io.EOF))
		fst.False(t, ev.CompareAndSwap(nil, io.EOF))
		fst.ErrorIs(t, ev.Load(), io.EOF)
		err1 := errors.New("hello")
		fst.ErrorIs(t, ev.Swap(err1), io.EOF)
		fst.ErrorIs(t, ev.Load(), err1)
		err2 := fmt.Errorf("world %w", io.EOF)
		ev.Store(err2)
		fst.ErrorIs(t, ev.Load(), err2)
	}
	t.Run("case 1", func(t *testing.T) {
		v1 := NewValue[error](nil)
		checkErrValue(t, v1)
	})
	t.Run("case 2", func(t *testing.T) {
		v1 := &Value[error]{}
		checkErrValue(t, v1)
	})

	t.Run("case 3", func(t *testing.T) {
		v1 := NewValue[error](nil)
		v1.Store(nil)
		fst.NoError(t, v1.Load())
		checkErrValue(t, v1)
	})
	t.Run("case 4", func(t *testing.T) {
		v1 := &Value[error]{}
		v1.Store(nil)
		fst.NoError(t, v1.Load())
		checkErrValue(t, v1)
	})
}

func TestValueInt(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		nv := NewValue(0)
		fst.Equal(t, 0, nv.Load())
		fst.True(t, nv.CompareAndSwap(0, 1))
		fst.Equal(t, 1, nv.Load())
		fst.Equal(t, 1, nv.Swap(2))
		fst.Equal(t, 2, nv.Load())
	})
	t.Run("case 2", func(t *testing.T) {
		nv := &Value[int]{}
		fst.Equal(t, 0, nv.Load())
		fst.True(t, nv.CompareAndSwap(0, 1))
		fst.Equal(t, 1, nv.Load())
		fst.Equal(t, 1, nv.Swap(2))
		fst.Equal(t, 2, nv.Load())
	})
}
