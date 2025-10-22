//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-01

package xctx

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestNewKey(t *testing.T) {
	key1 := NewKey()
	key2 := NewKey()
	xt.False(t, key1 == key2)
}
