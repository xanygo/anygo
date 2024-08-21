//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-21

package xlog

import (
	"context"
	"testing"

	"github.com/fsgo/fst"
)

func TestDefault(t *testing.T) {
	check := func(t *testing.T) {
		Error(context.Background(), "hello")
		Info(context.Background(), "hello")
		Debug(context.Background(), "hello")
		Warn(context.Background(), "hello")
	}
	fst.NotNil(t, Default())
	check(t)

	SetDefault(&NopLogger{})
	check(t)
}
