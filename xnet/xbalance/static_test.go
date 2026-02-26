//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-11

package xbalance

import (
	"context"
	"testing"

	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xt"
)

func TestStatic_Pick(t *testing.T) {
	as := NewStaticByAddr(xnet.NewAddr("tcp", "127.0.0.1:8080"), xnet.NewAddr("tcp", "127.0.0.2:8080"))
	for range 100 {
		node, err := as.Pick(context.Background())
		xt.NoError(t, err)
		xt.NotNil(t, node)
	}
}
