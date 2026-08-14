//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestDefault(t *testing.T) {
	check := func(t *testing.T) {
		Error(t.Context(), "hello")
		Info(t.Context(), "hello")
		Debug(t.Context(), "hello")
		Warn(t.Context(), "hello")
	}
	xt.NotNil(t, Default())
	check(t)

	SetDefault(&NopLogger{})
	check(t)
}
