//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-18

package xsync

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestValueErr(t *testing.T) {
	checkErrValue := func(t *testing.T, ev *Value[error]) {
		xt.True(t, ev.CompareAndSwap(nil, io.EOF))
		xt.False(t, ev.CompareAndSwap(nil, io.EOF))
		xt.ErrorIs(t, ev.Load(), io.EOF)
		err1 := errors.New("hello")
		xt.ErrorIs(t, ev.Swap(err1), io.EOF)
		xt.ErrorIs(t, ev.Load(), err1)
		err2 := fmt.Errorf("world %w", io.EOF)
		ev.Store(err2)
		xt.ErrorIs(t, ev.Load(), err2)
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
		xt.NoError(t, v1.Load())
		checkErrValue(t, v1)
	})
	t.Run("case 4", func(t *testing.T) {
		v1 := &Value[error]{}
		v1.Store(nil)
		xt.NoError(t, v1.Load())
		checkErrValue(t, v1)
	})
}

func TestValueInt(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		nv := NewValue(0)
		xt.Equal(t, 0, nv.Load())
		xt.True(t, nv.CompareAndSwap(0, 1))
		xt.Equal(t, 1, nv.Load())
		xt.Equal(t, 1, nv.Swap(2))
		xt.Equal(t, 2, nv.Load())
	})
	t.Run("case 2", func(t *testing.T) {
		nv := &Value[int]{}
		xt.Equal(t, 0, nv.Load())
		xt.True(t, nv.CompareAndSwap(0, 1))
		xt.Equal(t, 1, nv.Load())
		xt.Equal(t, 1, nv.Swap(2))
		xt.Equal(t, 2, nv.Load())
	})
}
