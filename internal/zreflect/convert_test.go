//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package zreflect_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestBaseTypeToString(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		str, ok := zreflect.BaseTypeToString(123)
		xt.True(t, ok)
		xt.Equal(t, "123", str)
	})

	t.Run("my-int", func(t *testing.T) {
		type myInt int
		str, ok := zreflect.BaseTypeToString(myInt(123))
		xt.True(t, ok)
		xt.Equal(t, "123", str)
	})
}

func TestBaseTypeToInt64(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		num, ok := zreflect.BaseTypeToInt64(123)
		xt.True(t, ok)
		xt.Equal(t, int64(123), num)
	})
	t.Run("my-int", func(t *testing.T) {
		type myInt int
		num, ok := zreflect.BaseTypeToInt64(myInt(123))
		xt.True(t, ok)
		xt.Equal(t, int64(123), num)
	})
}
