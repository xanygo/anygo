//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xsession

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestNewID(t *testing.T) {
	got := NewID()
	xt.Greater(t, len(got), idMinLen)
	xt.True(t, IsValidID(got))
}
